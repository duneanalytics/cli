#!/bin/sh
# Dune CLI installer
# Usage: curl -sSfL https://github.com/duneanalytics/cli/raw/main/install.sh | bash
#
# Environment variables:
#   INSTALL_DIR  — override installation directory (default: auto-detected)
#   VERSION      — specific version to install (default: latest)
#   GITHUB_TOKEN — GitHub token for private repo access

set -e

REPO="duneanalytics/cli"
BINARY="dune"
PROJECT="dune-cli"

main() {
    need_cmd uname
    need_cmd mktemp
    need_cmd chmod
    need_cmd rm

    os=$(detect_os)
    arch=$(detect_arch)
    version=$(resolve_version)

    if [ -z "$version" ]; then
        err "could not determine latest version"
    fi

    # Strip leading 'v' for archive name
    version_num="${version#v}"

    if [ -n "$INSTALL_DIR" ]; then
        install_dir="$INSTALL_DIR"
    else
        install_dir=$(detect_install_dir)
    fi

    case "$os" in
        windows) ext="zip" ;;
        *)       ext="tar.gz" ;;
    esac

    archive="${PROJECT}_${version_num}_${os}_${arch}.${ext}"
    url="https://github.com/${REPO}/releases/download/${version}/${archive}"
    checksum_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT

    log "Downloading ${BINARY} ${version} for ${os}/${arch}..."
    download "$url" "$tmp/$archive"
    download "$checksum_url" "$tmp/checksums.txt"

    log "Verifying checksum..."
    verify_checksum "$tmp/$archive" "$tmp/checksums.txt" "$archive"

    log "Extracting..."
    case "$ext" in
        tar.gz) tar -xzf "$tmp/$archive" -C "$tmp" ;;
        zip)    need_cmd unzip; unzip -q "$tmp/$archive" -d "$tmp" ;;
    esac

    binary_name="$BINARY"
    if [ "$os" = "windows" ]; then
        binary_name="${BINARY}.exe"
    fi

    if [ ! -f "$tmp/$binary_name" ]; then
        err "binary '$binary_name' not found in archive"
    fi

    chmod +x "$tmp/$binary_name"

    mkdir -p "$install_dir" 2>/dev/null || true

    if [ -w "$install_dir" ]; then
        mv "$tmp/$binary_name" "$install_dir/$binary_name"
    else
        log "Installing to ${install_dir} (requires sudo)..."
        sudo mkdir -p "$install_dir"
        sudo mv "$tmp/$binary_name" "$install_dir/$binary_name"
    fi

    log "Installed ${BINARY} ${version} to ${install_dir}/${binary_name}"
}

# Pick the best install directory by checking user-writable directories
# already on PATH, falling back to /usr/local/bin (always on PATH).
detect_install_dir() {
    for candidate in \
        "$HOME/.local/bin" \
        "$HOME/bin" \
        "$HOME/go/bin" \
        "$HOME/.cargo/bin"; do
        case ":$PATH:" in
            *":${candidate}:"*)
                if [ -d "$candidate" ] && [ -w "$candidate" ]; then
                    echo "$candidate"
                    return
                fi
                ;;
        esac
    done
    echo "/usr/local/bin"
}

detect_os() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux*)   echo "linux" ;;
        darwin*)  echo "darwin" ;;
        mingw*|msys*|cygwin*) echo "windows" ;;
        *)        err "unsupported OS: $os" ;;
    esac
}

detect_arch() {
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              err "unsupported architecture: $arch" ;;
    esac
}

resolve_version() {
    if [ -n "$VERSION" ]; then
        case "$VERSION" in
            v*) echo "$VERSION" ;;
            *)  echo "v$VERSION" ;;
        esac
        return
    fi

    auth_header=""
    if [ -n "$GITHUB_TOKEN" ]; then
        auth_header="Authorization: token $GITHUB_TOKEN"
    fi

    if has curl; then
        curl -sSfL -H "Accept: application/json" ${auth_header:+-H "$auth_header"} \
            "https://api.github.com/repos/${REPO}/releases/latest" \
            | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
    elif has wget; then
        wget -qO- --header="Accept: application/json" ${auth_header:+--header="$auth_header"} \
            "https://api.github.com/repos/${REPO}/releases/latest" \
            | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
    else
        err "need curl or wget to determine latest version"
    fi
}

download() {
    url="$1"
    dest="$2"

    if [ -n "$GITHUB_TOKEN" ]; then
        # Private repos require downloading via the API asset endpoint.
        # The direct github.com/.../releases/download/... URL returns 404.
        asset_name=$(basename "$url")
        release_url="https://api.github.com/repos/${REPO}/releases/tags/${version}"

        if has curl; then
            asset_api_url=$(curl -sSfL \
                -H "Authorization: token $GITHUB_TOKEN" \
                "$release_url" \
                | grep -B3 "\"name\": \"${asset_name}\"" \
                | sed -n 's/.*"url": "\(https:\/\/api\.github\.com\/repos\/.*\/releases\/assets\/[0-9]*\)".*/\1/p')

            if [ -n "$asset_api_url" ]; then
                curl -sSfL \
                    -H "Authorization: token $GITHUB_TOKEN" \
                    -H "Accept: application/octet-stream" \
                    -o "$dest" \
                    "$asset_api_url"
            else
                curl -sSfL -o "$dest" "$url"
            fi
        elif has wget; then
            asset_api_url=$(wget -qO- \
                --header="Authorization: token $GITHUB_TOKEN" \
                "$release_url" \
                | grep -B3 "\"name\": \"${asset_name}\"" \
                | sed -n 's/.*"url": "\(https:\/\/api\.github\.com\/repos\/.*\/releases\/assets\/[0-9]*\)".*/\1/p')

            if [ -n "$asset_api_url" ]; then
                wget -qO "$dest" \
                    --header="Authorization: token $GITHUB_TOKEN" \
                    --header="Accept: application/octet-stream" \
                    "$asset_api_url"
            else
                wget -qO "$dest" "$url"
            fi
        else
            err "need curl or wget to download files"
        fi
    else
        if has curl; then
            curl -sSfL -o "$dest" "$url"
        elif has wget; then
            wget -qO "$dest" "$url"
        else
            err "need curl or wget to download files"
        fi
    fi
}

verify_checksum() {
    file="$1"
    checksum_file="$2"
    archive_name="$3"

    expected=$(awk -v name="$archive_name" '$2 == name || $2 == "*"name { print $1; exit }' "$checksum_file")
    if [ -z "$expected" ]; then
        err "checksum not found for $archive_name"
    fi

    if has sha256sum; then
        actual=$(sha256sum "$file" | awk '{print $1}')
    elif has shasum; then
        actual=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        log "WARNING: could not verify checksum (no sha256sum or shasum found)"
        return 0
    fi

    if [ "$expected" != "$actual" ]; then
        err "checksum mismatch: expected $expected, got $actual"
    fi
}

has() {
    command -v "$1" > /dev/null 2>&1
}

need_cmd() {
    if ! has "$1"; then
        err "required command not found: $1"
    fi
}

log() {
    echo "  $*" >&2
}

err() {
    log "error: $*"
    exit 1
}

main "$@"

package main

import (
	"fmt"
	"os"

	"github.com/duneanalytics/cli/config"
)

func main() {
	env, err := config.FromEnvVars()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Printf("dune CLI initialized (host: %s)\n", env.Host)
}

package query

import (
	"fmt"
	"strconv"
)

func parseQueryID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid query ID %q: must be an integer", arg)
	}
	return id, nil
}

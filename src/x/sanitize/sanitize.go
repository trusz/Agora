package sanitize

import "github.com/microcosm-cc/bluemonday"

var p = bluemonday.NewPolicy()

func Sanitize(input string) string {
	return p.Sanitize(input)
}

func init() {
	// Initialize the package if needed
}

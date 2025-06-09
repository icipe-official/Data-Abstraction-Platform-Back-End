// Package json contains functions for manipulating values in their JSON form.
//
// Before function arguments are processed, they are converted to JSON and back into arguments of that map[string]any or []any.
package json

import (
	"fmt"
	"regexp"
	"strings"
)

// Regular expression to extract object keys and array indexes from path string into an array of strings using built-in string match function.
//
// Example: '8.childobject.array.[2][3].childobject'.match(new RegExp(/[^\.\[\]]+/, 'g')) results in the array ["8","childobject","array","2","3","childobject"]
func _KEY_ARRAY_INDEX_REGEX() *regexp.Regexp {
	return regexp.MustCompile(`[\.\[\]]+`)
}

func GetPathObjectKeyArrayIndexes(path string) []string {
	pathObjectKeyArrayIndexes := make([]string, 0)
	for _, value := range _KEY_ARRAY_INDEX_REGEX().Split(strings.Replace(path, "$.", "", 1), -1) {
		if strings.TrimSpace(value) != "" {
			pathObjectKeyArrayIndexes = append(pathObjectKeyArrayIndexes, value)
		}
	}
	return pathObjectKeyArrayIndexes
}

var ErrValueNotFoundError = fmt.Errorf("ValueNotFound")

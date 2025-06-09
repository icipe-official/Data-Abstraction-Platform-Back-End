package json

import (
	"fmt"
	"reflect"
	"strconv"
)

// Get value in object found at path.
//
// Parameters:
//
//   - object - Object or array to modify through deletion. Expected to be presented as if converted from JSON.
//
//   - path - Object-like path to value to get from object.
//     Numbers enclosed in square brackets or between full-stops indicate array indexes.
//
//     If path is empty, or equals to '$' then the object itself will be returned.
//
//     If path begins with `$.`, then it is removed. Inspired by Postgres' json path syntax.
//
//     Examples:
//
//     -- `$.[8].childobject.array[2][3].childobject`.
//
//     -- `$.8.childobject.array.2.3.childobject`.
//
// Returns value found or nil if not found.
func GetValueInObject(object any, path string) any {
	if len(path) == 0 || path == "$" {
		return object
	} else {
		return getValueInObject(object, GetPathObjectKeyArrayIndexes(path))
	}
}

func getValueInObject(currentValue any, pathObjectKeyArrayIndexes []string) any {
	if currentValue == nil || (reflect.TypeOf(currentValue).Kind() != reflect.Map && reflect.TypeOf(currentValue).Kind() != reflect.Slice) {
		return nil
	}
	typeOfCurrentValue := reflect.TypeOf(currentValue).Kind()

	currentPathKeyArrayIndex := func() any {
		if arrayIndex, err := strconv.Atoi(pathObjectKeyArrayIndexes[0]); err != nil {
			return pathObjectKeyArrayIndexes[0]
		} else {
			return arrayIndex
		}
	}()
	pathObjectKeyArrayIndexes = pathObjectKeyArrayIndexes[1:]

	switch reflect.TypeOf(currentPathKeyArrayIndex).Kind() {
	case reflect.String:
		return getValueInMap(currentValue, currentPathKeyArrayIndex, pathObjectKeyArrayIndexes)
	case reflect.Int:
		if typeOfCurrentValue == reflect.Slice {
			if currentPathKeyArrayIndex.(int) < len(currentValue.([]any)) {
				if len(pathObjectKeyArrayIndexes) > 0 {
					return getValueInObject(currentValue.([]any)[currentPathKeyArrayIndex.(int)], pathObjectKeyArrayIndexes)
				}
				return currentValue.([]any)[currentPathKeyArrayIndex.(int)]

			}
		} else {
			return getValueInMap(currentValue, currentPathKeyArrayIndex, pathObjectKeyArrayIndexes)
		}
	}
	return nil
}

func getValueInMap(currentValue any, currentPathKeyArrayIndex any, pathObjectKeyArrayIndexes []string) any {
	if reflect.TypeOf(currentValue).Kind() == reflect.Map {
		if value, ok := currentValue.(map[string]any)[fmt.Sprintf("%v", currentPathKeyArrayIndex)]; ok {
			if len(pathObjectKeyArrayIndexes) > 0 {
				return getValueInObject(value, pathObjectKeyArrayIndexes)
			}
			return value
		}
	}
	return nil
}

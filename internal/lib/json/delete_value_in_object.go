package json

import (
	"fmt"
	"reflect"
	"strconv"
)

// Delete value found at path in object.
//
// Parameters:
//
//   - object - Object or array to modify through deletion. Expected to be presented as if converted from JSON.
//
//   - path - Object-like path to value to remove from object.
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
// Returns object with value removed if it was found using the path.
func DeleteValueInObject(object any, path string) any {
	if len(path) == 0 || path == "$" {
		switch reflect.TypeOf(object).Kind() {
		case reflect.Map:
			return map[string]any{}
		case reflect.Slice:
			return []any{}
		default:
			return nil
		}
	}
	return deleteValueInObject(object, GetPathObjectKeyArrayIndexes(path))
}

func deleteValueInObject(currentValue any, pathObjectKeyArrayIndexes []string) any {
	if currentValue == nil || (reflect.TypeOf(currentValue).Kind() != reflect.Map && reflect.TypeOf(currentValue).Kind() != reflect.Slice) {
		return currentValue
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
		return deleteValueInMap(currentValue, currentPathKeyArrayIndex, pathObjectKeyArrayIndexes)
	case reflect.Int:
		if typeOfCurrentValue == reflect.Slice {
			if currentPathKeyArrayIndex.(int) > len(currentValue.([]any)) {
				return currentValue
			}
			if len(pathObjectKeyArrayIndexes) > 0 {
				currentValue.([]any)[currentPathKeyArrayIndex.(int)] = deleteValueInObject(currentValue.([]any)[currentPathKeyArrayIndex.(int)], pathObjectKeyArrayIndexes)
			} else {
				currentValue = append(currentValue.([]any)[:currentPathKeyArrayIndex.(int)], currentValue.([]any)[currentPathKeyArrayIndex.(int)+1:]...)
			}
		} else {
			return deleteValueInMap(currentValue, currentPathKeyArrayIndex, pathObjectKeyArrayIndexes)
		}
	}

	return currentValue
}

func deleteValueInMap(currentValue any, currentPathKeyArrayIndex any, pathObjectKeyArrayIndexes []string) any {
	if reflect.TypeOf(currentValue).Kind() == reflect.Map {
		if value, ok := currentValue.(map[string]any)[fmt.Sprintf("%v", currentPathKeyArrayIndex)]; ok {
			if len(pathObjectKeyArrayIndexes) > 0 {
				currentValue.(map[string]any)[fmt.Sprintf("%v", currentPathKeyArrayIndex)] = deleteValueInObject(value, pathObjectKeyArrayIndexes)
			} else {
				delete(currentValue.(map[string]any), fmt.Sprintf("%v", currentPathKeyArrayIndex))
			}
			return currentValue
		}
	}
	return currentValue
}

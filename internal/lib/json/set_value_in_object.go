package json

import (
	"reflect"
	"strconv"

	"github.com/brunoga/deep"
)

// Add or replace value in object with ValueToSet following the path.
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
//   - valueToSet - value to be added to object. Expected to be presented as if converted from JSON.
//
// Return Object with value added to it and error if converting valueToSet to Json and back failed.
func SetValueInObject(object any, path string, valueToSet any) (any, error) {
	valueToSetCopy, err := deep.Copy(valueToSet)
	if err != nil {
		return nil, err
	}

	var setValueInObject func(currentValue any, pathObjectKeyArrayIndexes []string) any
	setValueInObject = func(currentValue any, pathObjectKeyArrayIndexes []string) any {
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
			if currentValue == nil || reflect.TypeOf(currentValue).Kind() != reflect.Map {
				currentValue = map[string]any{}
			}
			if len(pathObjectKeyArrayIndexes) > 0 {
				currentValue.(map[string]any)[currentPathKeyArrayIndex.(string)] = setValueInObject(currentValue.(map[string]any)[currentPathKeyArrayIndex.(string)], pathObjectKeyArrayIndexes)
			} else {
				currentValue.(map[string]any)[currentPathKeyArrayIndex.(string)] = valueToSetCopy
			}
		case reflect.Int:
			if currentValue == nil || reflect.TypeOf(currentValue).Kind() != reflect.Slice {
				currentValue = []any{}
			}
			if currentPathKeyArrayIndex.(int) > len(currentValue.([]any))-1 {
				for i := len(currentValue.([]any)); i <= currentPathKeyArrayIndex.(int); i++ {
					currentValue = append(currentValue.([]any), nil)
				}
			}
			if len(pathObjectKeyArrayIndexes) > 0 {
				currentValue.([]any)[currentPathKeyArrayIndex.(int)] = setValueInObject(currentValue.([]any)[currentPathKeyArrayIndex.(int)], pathObjectKeyArrayIndexes)
			} else {
				currentValue.([]any)[currentPathKeyArrayIndex.(int)] = valueToSetCopy
			}
		default:
			return currentValue
		}

		return currentValue
	}

	if len(path) == 0 || path == "$" {
		if valueToSetCopy, err := deep.Copy(valueToSet); err != nil {
			return nil, err
		} else {
			return valueToSetCopy, nil
		}
	} else {
		return setValueInObject(object, GetPathObjectKeyArrayIndexes(path)), nil
	}
}

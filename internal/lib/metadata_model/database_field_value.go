package metadatamodel

import (
	"errors"
	"reflect"

	intlibjson "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/json"
)

func DatabaseGetColumnFieldValue(metadatamodel any, columnFieldName string, tableCollectionUID string, valueToGetFrom any, skipIfFGDisabled bool, skipIfDataExtraction bool) (any, error) {
	if len(columnFieldName) == 0 {
		return nil, FunctionNameAndError(DatabaseGetColumnFieldValue, errors.New("columnFieldName is empty"))
	}

	databaseColumnFields, err := DatabaseGetColumnFields(metadatamodel, tableCollectionUID, skipIfFGDisabled, skipIfDataExtraction)
	if err != nil {
		return nil, FunctionNameAndError(DatabaseGetColumnFieldValue, err)
	}

	if columnField, ok := databaseColumnFields.Fields[columnFieldName]; ok {
		pathToColumnFieldValue, err := GetValueAsString(columnField[FIELD_2D_POSITION_PROP_FIELD_GROUP_KEY])
		if err != nil {
			return nil, FunctionNameAndError(DatabaseGetColumnFieldValue, err)
		}
		pathToColumnFieldValue = GetPathToValue(pathToColumnFieldValue, true, "[0]")

		return intlibjson.GetValueInObject(valueToGetFrom, pathToColumnFieldValue), nil
	} else {
		return nil, FunctionNameAndError(DatabaseGetColumnFieldValue, errors.New("columnField is empty"))
	}
}

func DatabaseDeleteColumnFieldValue(metadatamodel any, columnFieldName string, tableCollectionUID string, valueToDeleteIn any, skipIfFGDisabled bool, skipIfDataExtraction bool) (any, error) {
	if len(columnFieldName) == 0 {
		return nil, FunctionNameAndError(DatabaseDeleteColumnFieldValue, errors.New("columnField is empty"))
	}

	databaseColumnFields, err := DatabaseGetColumnFields(metadatamodel, tableCollectionUID, skipIfFGDisabled, skipIfDataExtraction)
	if err != nil {
		return nil, FunctionNameAndError(DatabaseDeleteColumnFieldValue, err)
	}

	if columnField, ok := databaseColumnFields.Fields[columnFieldName]; ok {
		pathToColumnFieldValue, err := GetValueAsString(columnField[FIELD_2D_POSITION_PROP_FIELD_GROUP_KEY])
		if err != nil {
			return nil, FunctionNameAndError(DatabaseDeleteColumnFieldValue, err)
		}
		pathToColumnFieldValue = GetPathToValue(pathToColumnFieldValue, true, "[0]")

		return intlibjson.DeleteValueInObject(valueToDeleteIn, pathToColumnFieldValue), nil
	} else {
		return nil, FunctionNameAndError(DatabaseDeleteColumnFieldValue, errors.New("columnField is empty"))
	}
}

func DatabaseSetColumnFieldValue(metadatamodel any, columnFieldName string, tableCollectionUID string, valueToGetIn any, value any, skipIfFGDisabled bool, skipIfDataExtraction bool) (any, error) {
	if len(columnFieldName) == 0 {
		return nil, FunctionNameAndError(DatabaseSetColumnFieldValue, errors.New("columnField is empty"))
	}

	databaseColumnFields, err := DatabaseGetColumnFields(metadatamodel, tableCollectionUID, skipIfFGDisabled, skipIfDataExtraction)
	if err != nil {
		return nil, FunctionNameAndError(DatabaseSetColumnFieldValue, err)
	}

	if columnField, ok := databaseColumnFields.Fields[columnFieldName]; ok {
		pathToColumnFieldValue, err := GetValueAsString(columnField[FIELD_2D_POSITION_PROP_FIELD_GROUP_KEY])
		if err != nil {
			return nil, FunctionNameAndError(DatabaseSetColumnFieldValue, err)
		}
		pathToColumnFieldValue = GetPathToValue(pathToColumnFieldValue, true, "[0]")

		if reflect.TypeOf(value).Kind() == reflect.Slice || reflect.TypeOf(value).Kind() == reflect.Array {
			return intlibjson.SetValueInObject(valueToGetIn, pathToColumnFieldValue, value)
		}

		return intlibjson.SetValueInObject(valueToGetIn, pathToColumnFieldValue, []any{value})
	} else {
		return nil, FunctionNameAndError(DatabaseSetColumnFieldValue, errors.New("columnField is empty"))
	}
}

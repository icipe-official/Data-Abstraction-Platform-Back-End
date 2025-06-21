package metadatamodel

import (
	"errors"
	"fmt"

	"github.com/brunoga/deep"
)

type DatabaseColumnFields struct {
	ColumnFieldsReadOrder []string
	Fields                map[string]map[string]any
}

// Extracts database fields from metadatamodel if tableCollectionName matches.
//
// Parameters:
//
//   - metadatamodel - A valid metadata-model of type object (not array). Expected to presented as if converted from JSON.
//
//   - tableCollectionUID - Extract only fields whose FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID match this value.
//
//   - skipIfFGDisabled - Do not include field group if property FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE($FG_VIEW_DISABLE) is true.
//
//   - skipIfDataExtraction - Do not include field group if property FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION($DATABASE_SKIP_DATA_EXTRACTION) is true.
//
// returns error if metadatamodel or tableCollectionName is not valid.
func DatabaseGetColumnFields(metadatamodel any, tableCollectionUID string, skipIfFGDisabled bool, skipIfDataExtraction bool) (DatabaseColumnFields, error) {
	x, err := newdatabaseGetColumnFields(tableCollectionUID, skipIfFGDisabled, skipIfDataExtraction)
	if err != nil {
		return DatabaseColumnFields{}, FunctionNameAndError(DatabaseGetColumnFields, err)
	}

	if err := x.GetDatabaseColumnFields(metadatamodel); err != nil {
		return DatabaseColumnFields{}, FunctionNameAndError(DatabaseGetColumnFields, err)
	}
	return x.DatabaseColumnFields(), nil
}

type databaseGetColumnFields struct {
	databaseColumnFields DatabaseColumnFields
	tableCollectionUID   string
	skipIfFgDisabled     bool
	skipIfDataExtraction bool
}

func (n *databaseGetColumnFields) DatabaseColumnFields() DatabaseColumnFields {
	return n.databaseColumnFields
}

func newdatabaseGetColumnFields(tableCollectionUID string, skipIfFGDisabled bool, skipIfDataExtraction bool) (*databaseGetColumnFields, error) {
	n := new(databaseGetColumnFields)

	if len(tableCollectionUID) == 0 {
		return nil, FunctionNameAndError(newdatabaseGetColumnFields, errors.New("tableCollectionName is empty"))
	}
	n.tableCollectionUID = tableCollectionUID
	n.skipIfFgDisabled = skipIfFGDisabled
	n.skipIfDataExtraction = skipIfDataExtraction

	n.databaseColumnFields = DatabaseColumnFields{
		ColumnFieldsReadOrder: make([]string, 0),
		Fields:                make(map[string]map[string]any),
	}

	return n, nil
}

func (n *databaseGetColumnFields) GetDatabaseColumnFields(mmGroup any) error {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return FunctionNameAndError(n.GetDatabaseColumnFields, err)
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return FunctionNameAndError(n.GetDatabaseColumnFields, err)
	}

	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroup)
	if err != nil {
		return FunctionNameAndError(n.GetDatabaseColumnFields, err)
	}

	for _, fgKeySuffix := range mmGroupReadOrderOfFields {
		fgKeySuffixString, err := GetValueAsString(fgKeySuffix)
		if err != nil {
			return FunctionNameAndError(n.GetDatabaseColumnFields, err)
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeySuffixString])
		if err != nil {
			return FunctionNameAndError(n.GetDatabaseColumnFields, err)
		}

		if n.skipIfDataExtraction {
			if value, ok := fgMap[FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION].(bool); ok && value {
				continue
			}
		}

		if n.skipIfFgDisabled {
			if value, ok := fgMap[FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE].(bool); ok && value {
				continue
			}
		}

		if _, err := GetGroupFields(fgMap); err == nil {
			if value, ok := fgMap[FIELD_GROUP_PROP_GROUP_EXTRACT_AS_SINGLE_FIELD].(bool); !ok || !value {
				if _, err := GetGroupReadOrderOfFields(fgMap); err == nil {
					n.GetDatabaseColumnFields(fgMap)
					continue
				}
			}
		}

		if fgTableCollectionUid, ok := fgMap[FIELD_GROUP_PROP_DATABASE_TABLE_COLLECTION_UID].(string); !ok || fgTableCollectionUid != n.tableCollectionUID {
			continue
		}

		if fieldColumnName, ok := fgMap[FIELD_GROUP_PROP_DATABASE_FIELD_COLUMN_NAME].(string); ok && len(fieldColumnName) > 0 {
			if _, ok := n.databaseColumnFields.Fields[fieldColumnName]; ok {
				return FunctionNameAndError(n.GetDatabaseColumnFields, fmt.Errorf("duplciate fieldColumnName '%s' found", fieldColumnName))
			}

			var newField map[string]any = fgMap
			if value, err := deep.Copy(fgMap); err == nil {
				newField = value
			}

			n.databaseColumnFields.ColumnFieldsReadOrder = append(n.databaseColumnFields.ColumnFieldsReadOrder, fieldColumnName)
			n.databaseColumnFields.Fields[fieldColumnName] = newField
		} else {
			return FunctionNameAndError(n.GetDatabaseColumnFields, errors.New("fieldColumnName is empty"))
		}
	}

	return nil
}

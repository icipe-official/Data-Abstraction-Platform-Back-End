package metadatamodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/brunoga/deep"
)

type groupConversion struct {
	FieldGroupKey             string
	FgKeySuffix               string
	Fields2DIndexes           []int
	Fields2DPrimaryKeyIndexes []int
	Fields                    []fieldConversion
	Groups                    []groupConversion
}

type fieldConversion struct {
	FgKeySuffix                       string
	Fields2DIndexes                   []int
	FieldJoinSymbol                   string
	ColumnIndexesThatMatchIndexHeader [][]int
	ReadOrderOfFields                 []string
}

// Converts a 2D array into an array of objects following the metadata-model structure.
type Convert2DArrayToObjects struct {
	fields2D       []any
	objects        []any
	gConversion    groupConversion
	current2DArray [][]any
}

// Constructor for Convert2DArrayToObjects
//
// Parameters:
//
//   - metadatamodel - A valid metadata-model of type object (not array). Expected to presented as if converted from JSON.
//
//   - target2DFields - Current order of 2D fields.
//
//   - skipIfFGDisabled - Do not include field group if property FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE($FG_VIEW_DISABLE) is true.
//
//   - skipIfDataExtraction - Do not include field group if property FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION($DATABASE_SKIP_DATA_EXTRACTION) is true.
//
//   - databaseColumnNames - only pick fields with FIELD_GROUP_PROP_DATABASE_FIELD_COLUMN_NAME in this array. Match order as well.
//
// returns error if NewExtract2DFields or Convert2DArrayToObjects.initFgConversion returns an error.
func NewConvert2DArrayToObjects(metadatamodel any, target2DFields []any, skipIfFGDisabled bool, skipIfDataExtraction bool, databaseColumnNames []string) (*Convert2DArrayToObjects, error) {
	n := new(Convert2DArrayToObjects)
	if len(target2DFields) > 0 {
		n.fields2D = target2DFields
	} else {
		extract2DFields, err := NewExtract2DFields(metadatamodel, skipIfFGDisabled, skipIfDataExtraction, true, databaseColumnNames)
		if err != nil {
			return nil, FunctionNameAndError(NewConvert2DArrayToObjects, err)
		}
		extract2DFields.Extract()
		extract2DFields.Reposition()
		extract2DFields.RemoveSkipped()
		n.fields2D = extract2DFields.Fields()
	}

	if gConversion, err := n.initgConversion(metadatamodel, databaseColumnNames); err != nil {
		return nil, FunctionNameAndError(NewConvert2DArrayToObjects, err)
	} else {
		n.gConversion = gConversion
	}

	n.objects = make([]any, 0)

	return n, nil
}

func (n *Convert2DArrayToObjects) Objects() []any {
	return n.objects
}

func (n *Convert2DArrayToObjects) ResetObjects() {
	n.objects = make([]any, 0)
}

func (n *Convert2DArrayToObjects) initgConversion(mmGroup any, databaseColumnNames []string) (groupConversion, error) {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
	}

	mmGroupKeyString, err := GetValueAsString(mmGroupMap[FIELD_2D_POSITION_PROP_FIELD_GROUP_KEY])
	if err != nil {
		return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
	}

	mmGroupConversion := groupConversion{
		FieldGroupKey: mmGroupKeyString,
		FgKeySuffix: func() string {
			fgKeyStringSplit := strings.Split(mmGroupKeyString, ".")
			return fgKeyStringSplit[len(fgKeyStringSplit)-1]
		}(),
		Fields2DIndexes:           make([]int, 0),
		Fields2DPrimaryKeyIndexes: make([]int, 0),
		Groups:                    make([]groupConversion, 0),
		Fields:                    make([]fieldConversion, 0),
	}

	// Set 2D Fields Indexes for current mmGroup
	for fIndex, fValue := range n.fields2D {
		fValueMap, err := GetFieldGroupMap(fValue)
		if err != nil {
			return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
		}

		if len(databaseColumnNames) > 0 {
			if dbFieldColumnName, ok := fValueMap[FIELD_GROUP_PROP_DATABASE_FIELD_COLUMN_NAME].(string); ok {
				if !slices.Contains(databaseColumnNames, dbFieldColumnName) {
					continue
				}
			}
		}

		fValueMapKey, err := GetValueAsString(fValueMap[FIELD_GROUP_PROP_FIELD_GROUP_KEY])
		if err != nil {
			return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
		}

		relativePath := strings.Split(fValueMapKey, fmt.Sprintf("%s.%s%s.", mmGroupKeyString, FIELD_GROUP_PROP_GROUP_FIELDS, ARRAY_PATH_PLACEHOLDER))
		if len(relativePath) == 2 && len(strings.Split(relativePath[1], ".")) == 1 {
			mmGroupConversion.Fields2DIndexes = append(mmGroupConversion.Fields2DIndexes, fIndex)
		}
	}

	// Set 2D Fields Indexes that are primary keys
	if value, err := n.getPrimaryKey2DFieldsIndexes(mmGroup); err == nil && len(value) > 0 {
		mmGroupConversion.Fields2DPrimaryKeyIndexes = value
	} else {
		mmGroupConversion.Fields2DPrimaryKeyIndexes = mmGroupConversion.Fields2DIndexes
		if v, err := deep.Copy(mmGroupConversion.Fields2DIndexes); err == nil {
			mmGroupConversion.Fields2DPrimaryKeyIndexes = v
		}
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
	}

	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroup)
	if err != nil {
		return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
	}

	for _, fgKeySuffix := range mmGroupReadOrderOfFields {
		fgKeySuffixString, err := GetValueAsString(fgKeySuffix)
		if err != nil {
			return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeySuffixString])
		if err != nil {
			return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
		}

		if len(databaseColumnNames) > 0 {
			if dbFieldColumnName, ok := fgMap[FIELD_GROUP_PROP_DATABASE_FIELD_COLUMN_NAME].(string); ok {
				if !slices.Contains(databaseColumnNames, dbFieldColumnName) {
					continue
				}
			}
		}

		newFieldConversion := fieldConversion{
			FgKeySuffix: fgKeySuffixString,
		}

		if joinSymbol, ok := fgMap[FIELD_GROUP_PROP_FIELD_MULTIPLE_VALUES_JOIN_SYMBOL].(string); ok && len(joinSymbol) > 0 {
			newFieldConversion.FieldJoinSymbol = joinSymbol
		} else {
			newFieldConversion.FieldJoinSymbol = ","
		}

		fgMapKey, err := GetValueAsString(fgMap[FIELD_GROUP_PROP_FIELD_GROUP_KEY])
		if err != nil {
			return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
		}

		if _, err := GetGroupFields(fgMap); err == nil {
			if gReadOrderOfFields, err := GetGroupReadOrderOfFields(fgMap); err == nil {
				if value, ok := fgMap[FIELD_GROUP_PROP_GROUP_EXTRACT_AS_SINGLE_FIELD].(bool); ok && value {
					if value, err := n.get2DFieldsIndexesFromCurrentGroupIndexes(mmGroupConversion.Fields2DIndexes, fgMapKey); err != nil {
						return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
					} else if len(value) == 0 {
						return groupConversion{}, FunctionNameAndError(n.initgConversion, fmt.Errorf("Fields2DIndexes is empty | %s | %v", fgMapKey, mmGroupConversion.Fields2DIndexes))
					} else {
						newFieldConversion.Fields2DIndexes = value
						mmGroupConversion.Fields = append(mmGroupConversion.Fields, newFieldConversion)
					}
					continue
				}

				if fgViewMaxNoOfValuesInSeparateColumns := FgGet2DConversion(fgMap); fgViewMaxNoOfValuesInSeparateColumns > 1 {
					if g, err := func() ([]string, error) {
						gReadOrderOfFieldsString := make([]string, len(gReadOrderOfFields))
						for groofIndex, groof := range gReadOrderOfFields {
							if groofString, ok := groof.(string); ok {
								gReadOrderOfFieldsString[groofIndex] = groofString
							} else {
								return nil, argumentsError(n.initgConversion, "groof", "string", groof)
							}
						}
						return gReadOrderOfFieldsString, nil
					}(); err != nil {
						return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
					} else {
						newFieldConversion.ReadOrderOfFields = g
					}

					newFieldConversion.ColumnIndexesThatMatchIndexHeader = make([][]int, fgViewMaxNoOfValuesInSeparateColumns)
					for columnIndex := range fgViewMaxNoOfValuesInSeparateColumns {
						columnIndexHeaders := make([]int, 0)
						for fIndex, fValue := range n.fields2D {
							fMap, err := GetFieldGroupMap(fValue)
							if err != nil {
								return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
							}

							fKeyString, err := GetValueAsString(fMap[FIELD_GROUP_PROP_FIELD_GROUP_KEY])
							if err != nil {
								return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
							}

							if strings.HasPrefix(fKeyString, fgMapKey) {
								if fMap[FIELD_GROUP_PROP_FIELD_VIEW_VALUES_IN_SEPARATE_COLUMNS_HEADER_INDEX] == columnIndex {
									columnIndexHeaders = append(columnIndexHeaders, fIndex)
								}
							}
						}
						newFieldConversion.ColumnIndexesThatMatchIndexHeader[columnIndex] = columnIndexHeaders
					}

					mmGroupConversion.Fields = append(mmGroupConversion.Fields, newFieldConversion)
					continue
				}

				if gConversion, err := n.initgConversion(fgMap, databaseColumnNames); err != nil {
					return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
				} else {
					mmGroupConversion.Groups = append(mmGroupConversion.Groups, gConversion)
					continue
				}
			}
		}

		if value, err := n.get2DFieldsIndexesFromCurrentGroupIndexes(mmGroupConversion.Fields2DIndexes, fgMapKey); err != nil {
			return groupConversion{}, FunctionNameAndError(n.initgConversion, err)
		} else if len(value) == 0 {
			return groupConversion{}, FunctionNameAndError(n.initgConversion, fmt.Errorf("Fields2DIndexes is empty | %s | %v", fgMapKey, mmGroupConversion.Fields2DIndexes))
		} else {
			newFieldConversion.Fields2DIndexes = value
			mmGroupConversion.Fields = append(mmGroupConversion.Fields, newFieldConversion)
		}
	}

	return mmGroupConversion, nil
}

func (n *Convert2DArrayToObjects) Convert(data any) error {
	cleanedData := make([][]any, 0)
	if data2DArray, ok := data.([][]any); ok {
		for _, dValue := range data2DArray {
			includeRow := false
			for _, datumValue := range dValue {
				if datumValue != nil {
					includeRow = true
					break
				}
			}
			if includeRow {
				cleanedData = append(cleanedData, dValue)
			}
		}
	} else if dataArray, ok := data.([]any); ok {
		for _, dValue := range dataArray {
			if datumArray, ok := dValue.([]any); ok {
				includeRow := false
				for _, datumValue := range datumArray {
					if datumValue != nil {
						includeRow = true
						break
					}
				}
				if includeRow {
					cleanedData = append(cleanedData, datumArray)
				}
			} else {
				return argumentsError(n.Convert, "dValue", "[]any", data)
			}
		}
	} else {
		return argumentsError(n.Convert, "data", "[][]any", data)
	}

	currentDataIndexes := make([]int, len(cleanedData))
	for cdIndex := range cleanedData {
		currentDataIndexes[cdIndex] = cdIndex
	}

	n.current2DArray = cleanedData

	if newObjects, err := n.convert(n.gConversion, currentDataIndexes); err != nil {
		return FunctionNameAndError(n.Convert, err)
	} else if len(newObjects) > 0 {
		n.objects = append(n.objects, newObjects...)
	}

	return nil
}

func (n *Convert2DArrayToObjects) convert(gConversion groupConversion, current2DArrayRowIndexes []int) ([]any, error) {
	groupedDataIndexes, err := n.groupDataByPrimaryKeys(gConversion, current2DArrayRowIndexes)
	if err != nil {
		return nil, FunctionNameAndError(n.convert, err)
	}

	objects := make([]any, len(groupedDataIndexes))

	for gdIndex, gdValue := range groupedDataIndexes {
		object := make(map[string]any)

		for _, field := range gConversion.Fields {
			if len(field.ColumnIndexesThatMatchIndexHeader) > 0 && len(field.ReadOrderOfFields) > 0 {
				groupFieldValue := make([]any, 0)

				for _, ci2dhIndex := range field.ColumnIndexesThatMatchIndexHeader {
					newFieldValue := make(map[string]any)

					for citmihIndex, citmihIndexValue := range ci2dhIndex {
						fieldValue, err := n.extractFieldValueFromGroupedData(gdValue, []int{citmihIndexValue}, field.FieldJoinSymbol)
						if err != nil {
							return nil, FunctionNameAndError(n.convert, err)
						}

						if len(fieldValue) > 0 {
							newFieldValue[field.ReadOrderOfFields[citmihIndex]] = fieldValue
						}
					}

					if len(newFieldValue) > 0 {
						groupFieldValue = append(groupFieldValue, newFieldValue)
					}
				}

				if len(groupFieldValue) > 0 {
					object[field.FgKeySuffix] = groupFieldValue
				}

				continue
			}

			if len(field.Fields2DIndexes) > 0 {
				fieldValue, err := n.extractFieldValueFromGroupedData(gdValue, field.Fields2DIndexes, field.FieldJoinSymbol)
				if err != nil {
					return nil, FunctionNameAndError(n.convert, err)
				}
				if len(fieldValue) > 0 {
					object[field.FgKeySuffix] = fieldValue
				}
			}
		}

		if len(gConversion.Groups) > 0 {
			for _, gc := range gConversion.Groups {
				newObjectValue, err := n.convert(gc, gdValue)
				if err != nil {
					return nil, FunctionNameAndError(n.convert, err)
				}

				if len(newObjectValue) > 0 {
					object[gc.FgKeySuffix] = newObjectValue
				}
			}
		}

		if len(object) > 0 {
			objects[gdIndex] = object
		}
	}

	if len(objects) == 1 && objects[0] == nil {
		return nil, nil
	}

	newObjects := make([]any, 0)
	for _, aoo := range objects {
		if aoo != nil {
			newObjects = append(newObjects, aoo)
		}
	}
	return newObjects, nil
}

func (n *Convert2DArrayToObjects) extractFieldValueFromGroupedData(groupedDataIndexes []int, columnIndexes []int, joinSymbol string) ([]any, error) {
	duplicatedRowValues := make([][]any, 0)

	for _, gdIndex := range groupedDataIndexes {
		value := make([]any, 0)

		for _, cIndex := range columnIndexes {
			if n.current2DArray[gdIndex] == nil || cIndex >= len(n.current2DArray[gdIndex]) {
				return nil, FunctionNameAndError(n.extractFieldValueFromGroupedData, fmt.Errorf("n.current2DArray[gdIndex] and/or column at index %d not valid", cIndex))
			}

			if rowDataSlice, ok := n.current2DArray[gdIndex][cIndex].([]any); ok {
				value = append(value, rowDataSlice...)
			} else {
				value = append(value, n.current2DArray[gdIndex][cIndex])
			}
		}

		duplicatedRowValuesCopy := make([][]any, len(duplicatedRowValues))
		for i, v := range duplicatedRowValues {
			duplicatedRowValuesCopy[i] = make([]any, len(v))
			copy(duplicatedRowValuesCopy[i], v)
		}
		duplicatedRowValues = append(duplicatedRowValuesCopy, value)
	}

	if len(duplicatedRowValues) > 0 {
		for _, drValue := range duplicatedRowValues {
			if !reflect.DeepEqual(duplicatedRowValues[0], drValue) {
				return nil, FunctionNameAndError(n.extractFieldValueFromGroupedData, fmt.Errorf("duplicateRowValues not equal"))
			}
		}
	}

	duplicatedRowIsEmpty := true
	for _, drValue := range duplicatedRowValues[0] {
		if drValueString, ok := drValue.(string); ok {
			if len(drValueString) > 0 {
				duplicatedRowIsEmpty = false
				break
			}
		}
		if drValue != nil {
			duplicatedRowIsEmpty = false
			break
		}
	}

	if duplicatedRowIsEmpty {
		return []any{}, nil
	}

	if len(duplicatedRowValues[0]) == 1 {
		if drvString, ok := duplicatedRowValues[0][0].(string); ok {
			new := []any{}
			for _, v := range strings.Split(drvString, joinSymbol) {
				var newValue any
				if err := json.Unmarshal([]byte(v), &newValue); err != nil {
					newValue = v
				}
				new = append(new, newValue)
			}
			return new, nil
		}
	}

	return duplicatedRowValues[0], nil
}

func (n *Convert2DArrayToObjects) groupDataByPrimaryKeys(gConversion groupConversion, current2DArrayRowIndexes []int) ([][]int, error) {
	groupedDataIndexes := make([][]int, 0)
	currentDataIndexesProcessed := make([]int, 0)
	currentGroupDataIndexesRow := 0

	for _, cdIndex := range current2DArrayRowIndexes {
		if slices.Contains(currentDataIndexesProcessed, cdIndex) {
			continue
		}
		currentDataIndexesProcessed = append(currentDataIndexesProcessed, cdIndex)

		groupedDataIndexes = append(groupedDataIndexes, []int{cdIndex})
		for _, compCdIndex := range current2DArrayRowIndexes {
			if slices.Contains(currentDataIndexesProcessed, compCdIndex) {
				continue
			}

			cdPrimaryKeyValues, err := n.getPrimaryKeysValuesFromDataRow(gConversion.Fields2DPrimaryKeyIndexes, cdIndex)
			if err != nil {
				return nil, FunctionNameAndError(n.groupDataByPrimaryKeys, err)
			}

			compCdPrimaryKeyValues, err := n.getPrimaryKeysValuesFromDataRow(gConversion.Fields2DPrimaryKeyIndexes, compCdIndex)
			if err != nil {
				return nil, FunctionNameAndError(n.groupDataByPrimaryKeys, err)
			}

			if reflect.DeepEqual(cdPrimaryKeyValues, compCdPrimaryKeyValues) {
				currentDataIndexesProcessed = append(currentDataIndexesProcessed, compCdIndex)
				groupedDataIndexes[currentGroupDataIndexesRow] = append(groupedDataIndexes[currentGroupDataIndexesRow], compCdIndex)
			}
		}

		currentGroupDataIndexesRow += 1
	}

	return groupedDataIndexes, nil
}

func (n *Convert2DArrayToObjects) getPrimaryKeysValuesFromDataRow(fieldPrimaryKeysIndexes []int, dataRowIndex int) ([]any, error) {
	primaryKeysValues := make([]any, 0)

	for _, fcpkIndex := range fieldPrimaryKeysIndexes {
		if n.current2DArray[dataRowIndex] != nil && fcpkIndex < len(n.current2DArray[dataRowIndex]) {
			primaryKeysValues = append(primaryKeysValues, n.current2DArray[dataRowIndex][fcpkIndex])
			continue
		}
		return nil, FunctionNameAndError(n.getPrimaryKeysValuesFromDataRow, errors.New("n.current2DArray[dataRowIndex] is not valid"))
	}

	if jsonData, err := json.Marshal(primaryKeysValues); err != nil {
		return nil, FunctionNameAndError(n.getPrimaryKeysValuesFromDataRow, fmt.Errorf("convert primaryKeysValues to json failed, error: %v", err))
	} else {
		var jsonParsed []any
		if err := json.Unmarshal(jsonData, &jsonParsed); err != nil {
			return nil, FunctionNameAndError(n.getPrimaryKeysValuesFromDataRow, fmt.Errorf("convert primaryKeysValues from json failed, error: %v", err))
		}
		return jsonParsed, nil
	}
}

func (n *Convert2DArrayToObjects) get2DFieldsIndexesFromCurrentGroupIndexes(currentGroupIndexes []int, fgKey string) ([]int, error) {
	fields2DIndexes := make([]int, 0)

	for _, cgIndex := range currentGroupIndexes {
		cgMap, err := GetFieldGroupMap(n.fields2D[cgIndex])
		if err != nil {
			return nil, FunctionNameAndError(n.get2DFieldsIndexesFromCurrentGroupIndexes, err)
		}

		if cgMap[FIELD_GROUP_PROP_FIELD_GROUP_KEY] == fgKey {
			fields2DIndexes = append(fields2DIndexes, cgIndex)
		}
	}

	return fields2DIndexes, nil
}

func (n *Convert2DArrayToObjects) getPrimaryKey2DFieldsIndexes(mmGroup any) ([]int, error) {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return nil, FunctionNameAndError(n.getPrimaryKey2DFieldsIndexes, err)
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return nil, FunctionNameAndError(n.getPrimaryKey2DFieldsIndexes, err)
	}

	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroupMap)
	if err != nil {
		return nil, FunctionNameAndError(n.getPrimaryKey2DFieldsIndexes, err)
	}

	primaryKeyIndexes := make([]int, 0)

	for _, fgKey := range mmGroupReadOrderOfFields {
		fgKeyString, err := GetValueAsString(fgKey)
		if err != nil {
			return nil, FunctionNameAndError(n.getPrimaryKey2DFieldsIndexes, err)
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeyString])
		if err != nil {
			return nil, FunctionNameAndError(n.getPrimaryKey2DFieldsIndexes, err)
		}

		if value, ok := fgMap[FIELD_GROUP_PROP_FIELD_GROUP_IS_PRIMARY_KEY].(bool); ok && value {
			if _, err := GetGroupFields(fgMap); err == nil {
				if _, err := GetGroupReadOrderOfFields(fgMap); err == nil {
					if pKeyIndexes, err := n.getPrimaryKey2DFieldsIndexes(fgMap); err != nil {
						return nil, FunctionNameAndError(n.getPrimaryKey2DFieldsIndexes, err)
					} else {
						primaryKeyIndexes = append(primaryKeyIndexes, pKeyIndexes...)
					}

				}
			}

			for fIndex, fValue := range n.fields2D {
				fValueMap, err := GetFieldGroupMap(fValue)
				if err != nil {
					return nil, FunctionNameAndError(n.getPrimaryKey2DFieldsIndexes, err)
				}

				if fgMap[FIELD_GROUP_PROP_FIELD_GROUP_KEY] == fValueMap[FIELD_GROUP_PROP_FIELD_GROUP_KEY] {
					primaryKeyIndexes = append(primaryKeyIndexes, fIndex)
				}
			}
		}
	}

	return primaryKeyIndexes, nil
}

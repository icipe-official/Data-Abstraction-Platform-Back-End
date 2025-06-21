package metadatamodel

import (
	"errors"
	"fmt"
	"slices"

	"github.com/brunoga/deep"
)

// Extracts fields that can be used for working with data in 2D array form like in excel or csv.
//
// Recommended order to Extract fields:
//
// - Extract2DFields.Extract
//
// - Extract2DFields.Reposition
//
// - Extract2DFields.RemoveSkipped
//
// - Extract2DFields.Fields
type Extract2DFields struct {
	fields               []any
	skipIfFgDisabled     bool
	skipIfDataExtraction bool
	removePrimaryKey     bool
	metadataModel        map[string]any
	repositionFields     RepositionFields
	databaseColumnNames  []string
}

// Constructor for Extract2DFields
//
// Parameters:
//
//   - metadatamodel - A valid metadata-model of type object (not array). Expected to presented as if converted from JSON.
//
//   - skipIfFGDisabled - Do not include field group if property FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE($FG_VIEW_DISABLE) is true.
//
//   - skipIfDataExtraction - Do not include field group if property FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION($DATABASE_SKIP_DATA_EXTRACTION) is true.
//
//   - removePrimaryKey - Remove field if FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE or FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION even if FIELD_GROUP_PROP_FIELD_GROUP_IS_PRIMARY_KEY field property is true.
//
//   - databaseColumnNames - only pick fields with FIELD_GROUP_PROP_DATABASE_FIELD_COLUMN_NAME in this array. Match order as well.
//
// returns error if argument metadatamodel is not of type map[string]any.
func NewExtract2DFields(metadatamodel any, skipIfFGDisabled bool, skipIfDataExtraction bool, removePrimaryKey bool, databaseColumnNames []string) (*Extract2DFields, error) {
	n := new(Extract2DFields)
	if metadatamodelMap, ok := metadatamodel.(map[string]any); ok {
		n.metadataModel = metadatamodelMap
	} else {
		return nil, FunctionNameAndError(NewExtract2DFields, errors.New("argument metadatamodelMap is not of type map[string]any"))
	}
	n.skipIfFgDisabled = skipIfFGDisabled
	n.skipIfDataExtraction = skipIfDataExtraction
	n.removePrimaryKey = removePrimaryKey
	n.databaseColumnNames = databaseColumnNames
	return n, nil
}

// Return information about fields that need to be repositioned based on FIELD_GROUP_PROP_FIELD_2D_VIEW_POSITION when Extract2DFields.Extract is called.
func (n *Extract2DFields) RepositionFields() RepositionFields {
	return n.repositionFields
}

// Return Extract2DFields.fields in its current state.
func (n *Extract2DFields) Fields() []any {
	return n.fields
}

// Extracts fields found in  Extract2DFields.metadataModel and places them in Extract2DFields.fields.
func (n *Extract2DFields) Extract() error {
	if err := n.extract(n.metadataModel, false, false); err != nil {
		return FunctionNameAndError(n.Extract, err)
	}

	if len(n.databaseColumnNames) > 0 {
		newFields := make([]any, 0)
		for _, dbFieldColumnName := range n.databaseColumnNames {
			for _, field := range n.fields {
				if fieldMap, ok := field.(map[string]any); ok {
					if dbFCName, ok := fieldMap[FIELD_GROUP_PROP_DATABASE_FIELD_COLUMN_NAME].(string); ok {
						if dbFCName == dbFieldColumnName {
							newFields = append(newFields, field)
						}
					}
				}
			}
		}
		if len(newFields) == len(n.databaseColumnNames) {
			n.fields = newFields
		}
	}

	return nil
}

// Recursive function that extracts fields found in mmGroup and places them in Extract2DFields.fields.
//
// Will not remove any fields if FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION or FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE is true.
//
// Parameters:
//
// - mmGroup - Current metadata model Group to extract fields from.
//
// - mmGroupSkipDataExtraction - Add FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION property to child fields and groups of mmGroup if true.
//
// - mmGroupViewDisable - Add FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE property to child fields and groups of mmGroup if true.
//
// mm - Alias for metadata model.
//
// fg - Alias for field group.
func (n *Extract2DFields) extract(mmGroup any, mmGroupSkipDataExtraction bool, mmGroupViewDisable bool) error {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return FunctionNameAndError(n.extract, err)
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return FunctionNameAndError(n.extract, err)
	}

	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroup)
	if err != nil {
		return FunctionNameAndError(n.extract, err)
	}

	for _, fgKeySuffix := range mmGroupReadOrderOfFields {
		fgKeySuffixString, err := GetValueAsString(fgKeySuffix)
		if err != nil {
			return FunctionNameAndError(n.extract, err)
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeySuffixString])
		if err != nil {
			return FunctionNameAndError(n.extract, err)
		}

		if fgGroupFields, err := GetGroupFields(fgMap); err == nil {
			if fgGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(fgMap); err == nil {
				skipDataExtraction := mmGroupSkipDataExtraction
				if value, ok := fgMap[FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION].(bool); ok {
					skipDataExtraction = value
				}
				viewDisable := mmGroupViewDisable
				if value, ok := fgMap[FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE].(bool); ok {
					viewDisable = value
				}

				if value, ok := fgMap[FIELD_GROUP_PROP_GROUP_EXTRACT_AS_SINGLE_FIELD].(bool); ok && value {
					var newField map[string]any = fgMap
					if value, err := deep.Copy(fgMap); err == nil {
						newField = value
					}

					n.appendField(newField)
					continue
				}

				if fgViewMaxNoOfValuesInSeparateColumns := FgGet2DConversion(fgMap); fgViewMaxNoOfValuesInSeparateColumns > 1 {
					for columnIndex := range fgViewMaxNoOfValuesInSeparateColumns {
						for _, gfKey := range fgGroupReadOrderOfFields {
							gfKeyString, err := GetValueAsString(gfKey)
							if err != nil {
								return FunctionNameAndError(n.extract, err)
							}

							gfMap, err := GetFieldGroupMap(fgGroupFields[gfKeyString])
							if err != nil {
								return FunctionNameAndError(n.extract, err)
							}

							var newField map[string]any = gfMap
							if value, err := deep.Copy(gfMap); err == nil {
								newField = value
							}

							n.updateSeparateColumnsField(newField, mmGroupSkipDataExtraction, mmGroupViewDisable, columnIndex)

							n.appendField(newField)
						}
					}
					continue
				}

				n.extract(fgMap, skipDataExtraction, viewDisable)
				continue
			}
		}

		if fgViewMaxNoOfValuesInSeparateColumns := FgGet2DConversion(fgMap); fgViewMaxNoOfValuesInSeparateColumns > 1 {
			for columnIndex := range fgViewMaxNoOfValuesInSeparateColumns {
				var newField map[string]any = fgMap
				if value, err := deep.Copy(newField); err == nil {
					newField = value
				}

				n.updateSeparateColumnsField(newField, mmGroupSkipDataExtraction, mmGroupViewDisable, columnIndex)

				n.appendField(newField)
			}

			continue
		}

		var newField map[string]any = fgMap
		if value, err := deep.Copy(fgMap); err == nil {
			newField = value
		}

		if value := Get2DFieldViewPosition(newField); value != nil {
			n.repositionFields[len(n.fields)] = value
		}

		if mmGroupSkipDataExtraction {
			newField[FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION] = true
		}

		if mmGroupViewDisable {
			newField[FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE] = true
		}

		n.appendField(newField)
	}

	return nil
}

func (n *Extract2DFields) appendField(field map[string]any) {
	if len(n.databaseColumnNames) > 0 {
		if dbFieldColumnName, ok := field[FIELD_GROUP_PROP_DATABASE_FIELD_COLUMN_NAME].(string); ok {
			if slices.Contains(n.databaseColumnNames, dbFieldColumnName) {
				n.fields = append(n.fields, field)
			}
		}
	} else {
		n.fields = append(n.fields, field)
	}
}

func (n *Extract2DFields) updateSeparateColumnsField(field map[string]any, mmGroupSkipDataExtraction bool, mmGroupViewDisable bool, columnIndex int) {
	field[FIELD_2D_POSITION_PROP_FIELD_VIEW_VALUES_IN_SEPARATE_COLUMNS_HEADER_INDEX] = columnIndex
	if mmGroupSkipDataExtraction {
		field[FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION] = true
	}

	if mmGroupViewDisable {
		field[FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE] = true
	}

	if fgHeaderFormat, ok := field[FIELD_GROUP_PROP_FIELD_VIEW_VALUES_IN_SEPARATE_COLUMNS_HEADER_FORMAT].(string); ok && len(fgHeaderFormat) > 0 {
		field[FIELD_GROUP_PROP_FIELD_GROUP_NAME] = string(ARRAY_PATH_REGEX_SEARCH().ReplaceAll([]byte(fgHeaderFormat), fmt.Appendf(nil, "%d", columnIndex+1)))
	} else {
		field[FIELD_GROUP_PROP_FIELD_GROUP_NAME] = fmt.Sprintf("%s %v", GetFieldGroupName(field, "#unnamed"), columnIndex+1)
	}
}

// Calls Reposition2DFields and updates Extract2DFields.fields.
//
// Returns an error if Reposition2DFields returns an error.
func (n *Extract2DFields) Reposition() error {
	if newFields, err := Reposition2DFields(Field2DsWithReposition{Fields: n.fields, Reposition: n.repositionFields}); err != nil {
		return FunctionNameAndError(n.Reposition, err)
	} else {
		n.fields = newFields
	}

	return nil
}

// Returns result of calling Reposition2DFields.
func (n *Extract2DFields) FieldsRepositioned() ([]any, error) {
	return Reposition2DFields(Field2DsWithReposition{Fields: n.fields, Reposition: n.repositionFields})
}

// Calls RemoveSkipped2DFields and updates Extract2DFields.fields.
//
// Returns an error if RemoveSkipped2DFields returns an error.
func (n *Extract2DFields) RemoveSkipped() error {
	if newFields, err := RemoveSkipped2DFields(n.fields, n.skipIfFgDisabled, n.skipIfDataExtraction, n.removePrimaryKey); err != nil {
		return FunctionNameAndError(n.RemoveSkipped, err)
	} else {
		n.fields = newFields
	}

	return nil
}

func (n *Extract2DFields) FieldsWithSkippedRemoved() ([]any, error) {
	return RemoveSkipped2DFields(n.fields, n.skipIfFgDisabled, n.skipIfDataExtraction, n.removePrimaryKey)
}

// Reorder columns of each row in data to match order in targetFields.
type Reorder2DFields struct {
	sourceToTargetReadOrderOfFields []int
	targetToSourceReadOrderOfFields []int
}

// Constructor for Reorder2DFields.
//
// Initializes Reorder2DFields.sourceToTargetReadOrderOfFields and Reorder2DFields.targetToSourceReadOrderOfFields.
//
// Parameters:
//
//   - sourceFields - Current order of columns in each row of data.
//
//   - skipIfFGDisabled - Target order of columns in each row of data.
//
// returns error if sourceFields and targetFields are not valid.
func NewReorder2DFields(sourceFields []any, targetFields []any) (*Reorder2DFields, error) {
	n := new(Reorder2DFields)
	if len(sourceFields) != len(targetFields) {
		return nil, FunctionNameAndError(NewReorder2DFields, errors.New("length of sourceFields and targetFields not equal"))
	}

	n.sourceToTargetReadOrderOfFields = make([]int, 0)
	n.targetToSourceReadOrderOfFields = make([]int, len(sourceFields))

	for tfIndex, tfValue := range targetFields {
		sourceIndex := -1

		tfValueMap, err := GetFieldGroupMap(tfValue)
		if err != nil {
			return nil, FunctionNameAndError(NewReorder2DFields, err)
		}

		for sfIndex, sfValue := range sourceFields {
			sfValueMap, err := GetFieldGroupMap(sfValue)
			if err != nil {
				return nil, FunctionNameAndError(NewReorder2DFields, err)
			}

			if sfValueMap[FIELD_2D_POSITION_PROP_FIELD_GROUP_KEY] == tfValueMap[FIELD_2D_POSITION_PROP_FIELD_GROUP_KEY] {
				if value := Get2DFieldViewPosition(sfValueMap); value != nil {
					if value.FViewValuesInSeparateColumnsHeaderIndex != nil {
						if tfValueMap[FIELD_GROUP_PROP_FIELD_VIEW_VALUES_IN_SEPARATE_COLUMNS_HEADER_INDEX] != *value.FViewValuesInSeparateColumnsHeaderIndex {
							continue
						}
					}
				}
				n.sourceToTargetReadOrderOfFields = append(n.sourceToTargetReadOrderOfFields, sfIndex)
				sourceIndex = sfIndex
				break
			}
		}

		if sourceIndex < 0 {
			return nil, FunctionNameAndError(NewReorder2DFields, errors.New("targetField not found in sourceField"))
		}

		n.targetToSourceReadOrderOfFields[sourceIndex] = tfIndex
	}

	return n, nil
}

// Reorder columns in each row in data. Will modify order of columns in rows of original data so deep copy may needed if original order is required.
//
// returns an error if length of datum in data and Reorder2DFields.sourceToTargetReadOrderOfFields are not equal.
func (n *Reorder2DFields) Reorder(data [][]any) error {
	for dIndex, dValue := range data {
		if len(dValue) != len(n.sourceToTargetReadOrderOfFields) {
			return FunctionNameAndError(n.Reorder, fmt.Errorf("length of datum at index %d and sourceToTargetReadOrderOfFields not equal", dIndex))
		}

		targetDatum := make([]any, len(n.sourceToTargetReadOrderOfFields))
		for stIndex, stValue := range n.sourceToTargetReadOrderOfFields {
			tDatumData := data[dIndex][stValue]
			if value, err := deep.Copy(data[dIndex][stValue]); err == nil {
				tDatumData = value
			}
			targetDatum[stIndex] = tDatumData
		}

		data[dIndex] = targetDatum
	}

	return nil
}

// Remove 2D fields if FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE($FG_VIEW_DISABLE) or FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION($DATABASE_SKIP_DATA_EXTRACTION) is true.
//
// Returns updated fields and an error field is invalid.
func RemoveSkipped2DFields(fields []any, skipIfFgDisabled bool, skipIfDataExtraction bool, removePrimaryKey bool) ([]any, error) {
	newFields := []any{}

	for _, fValue := range fields {
		fValueMap, err := GetFieldGroupMap(fValue)
		if err != nil {
			return nil, FunctionNameAndError(RemoveSkipped2DFields, err)
		}

		if skipIfFgDisabled {
			if value, ok := fValueMap[FIELD_GROUP_PROP_FIELD_GROUP_VIEW_DISABLE].(bool); ok && value {
				if value, ok := fValueMap[FIELD_GROUP_PROP_FIELD_GROUP_IS_PRIMARY_KEY].(bool); ok && value {
					if removePrimaryKey {
						continue
					}
				} else {
					continue
				}
			}
		}

		if skipIfDataExtraction {
			if value, ok := fValueMap[FIELD_GROUP_PROP_DATABASE_SKIP_DATA_EXTRACTION].(bool); ok && value {
				if value, ok := fValueMap[FIELD_GROUP_PROP_FIELD_GROUP_IS_PRIMARY_KEY].(bool); ok && value {
					if removePrimaryKey {
						continue
					}
				} else {
					continue
				}
			}
		}

		newField := fValue
		if value, err := deep.Copy(fValue); err == nil {
			newField = value
		}
		newFields = append(newFields, newField)
	}

	return newFields, nil
}

// Rearrange 2D fields based on fields with FIELD_GROUP_PROP_FIELD_2D_VIEW_POSITION extracted during Extract2DFields.
//
// Returns error if deep copy of fields2D.Fields fails or repositionedFields is not valid.
func Reposition2DFields(fields2D Field2DsWithReposition) ([]any, error) {
	var repositionedFields []any
	if value, err := deep.Copy(fields2D.Fields); err != nil {
		return nil, FunctionNameAndError(Reposition2DFields, err)
	} else {
		repositionedFields = value
	}

	for sourceIndex, rValue := range fields2D.Reposition {
		destinationIndex := -1

		for rfIndex, rfValue := range repositionedFields {
			rfValueMap, err := GetFieldGroupMap(rfValue)
			if err != nil {
				continue
			}

			if rfValueMap[FIELD_2D_POSITION_PROP_FIELD_GROUP_KEY] == rValue.FgKey {
				if rValue.FViewValuesInSeparateColumnsHeaderIndex != nil {
					if rfValueMap[FIELD_GROUP_PROP_FIELD_VIEW_VALUES_IN_SEPARATE_COLUMNS_HEADER_INDEX] != *rValue.FViewValuesInSeparateColumnsHeaderIndex {
						continue
					}
				}

				if rValue.FieldPositionBefore != nil && *rValue.FieldPositionBefore {
					destinationIndex = rfIndex
				} else {
					destinationIndex = rfIndex + 1
				}
			}
		}

		if destinationIndex >= 0 {
			sourceCopy := deep.MustCopy(repositionedFields[sourceIndex])
			repositionedFields = append(repositionedFields[0:sourceIndex], repositionedFields[sourceIndex+1:]...)
			repositionedFields = append(repositionedFields[0:destinationIndex], append([]any{sourceCopy}, repositionedFields[destinationIndex:]...)...)
		}
	}

	return repositionedFields, nil
}

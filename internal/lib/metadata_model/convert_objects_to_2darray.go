package metadatamodel

import (
	"errors"
	"fmt"

	"github.com/brunoga/deep"
	intlibjson "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/json"
)

type fieldGroupConversion struct {
	FgKey                  *string
	FgSepColsMaxValues     *int
	FieldJoinSymbol        string
	FieldGroups            []fieldGroupConversion
	GroupReadOrderOfFields []string
}

// Converts an object or array of objects into a 2D array following the metadata-model structure.
type ConvertObjectsTo2DArray struct {
	array2D              [][]any
	skipIfFgDisabled     bool
	skipIfDataExtraction bool
	fgConversion         fieldGroupConversion
	currentDatum         map[string]any
	reorder2DFields      *Reorder2DFields
}

// Constructor for ConvertObjectsTo2DArray
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
// returns error if NewExtract2DFields or Convert2DArrayToObjects.initFgConversion returns an error.
func NewConvertObjectsTo2DArray(metadatamodel any, target2DFields []any, skipIfFGDisabled bool, skipIfDataExtraction bool) (*ConvertObjectsTo2DArray, error) {
	n := new(ConvertObjectsTo2DArray)
	n.skipIfFgDisabled = skipIfFGDisabled
	n.skipIfDataExtraction = skipIfDataExtraction

	if fgConversion, err := n.initFgConversion(metadatamodel); err != nil {
		return nil, FunctionNameAndError(NewConvertObjectsTo2DArray, err)
	} else {
		n.fgConversion = fgConversion
	}

	extract2DFields, err := NewExtract2DFields(metadatamodel, skipIfFGDisabled, skipIfDataExtraction, true, nil)
	if err != nil {
		return nil, FunctionNameAndError(NewConvertObjectsTo2DArray, err)
	}
	extract2DFields.Extract()
	dataFields, err := extract2DFields.FieldsWithSkippedRemoved()
	if err != nil {
		return nil, FunctionNameAndError(NewConvertObjectsTo2DArray, err)
	}

	if len(target2DFields) > 0 {
		if reorder2DFields, err := NewReorder2DFields(dataFields, target2DFields); err != nil {
			return nil, FunctionNameAndError(NewConvertObjectsTo2DArray, err)
		} else {
			n.reorder2DFields = reorder2DFields
		}
	} else {
		if len(extract2DFields.RepositionFields()) > 0 {
			extract2DFields.Reposition()
			extract2DFields.RemoveSkipped()
			if reorder2DFields, err := NewReorder2DFields(dataFields, extract2DFields.Fields()); err != nil {
				return nil, FunctionNameAndError(NewConvertObjectsTo2DArray, err)
			} else {
				n.reorder2DFields = reorder2DFields
			}
		}
	}

	return n, nil
}

func (n *ConvertObjectsTo2DArray) Array2D() [][]any {
	return n.array2D
}

func (n *ConvertObjectsTo2DArray) ResetArray2D() {
	n.array2D = make([][]any, 0)
}

func (n *ConvertObjectsTo2DArray) initFgConversion(mmGroup any) (fieldGroupConversion, error) {
	mmGroupMap, err := GetFieldGroupMap(mmGroup)
	if err != nil {
		return fieldGroupConversion{}, FunctionNameAndError(n.initFgConversion, err)
	}

	mmGroupFields, err := GetGroupFields(mmGroupMap)
	if err != nil {
		return fieldGroupConversion{}, FunctionNameAndError(n.initFgConversion, err)
	}

	mmGroupReadOrderOfFields, err := GetGroupReadOrderOfFields(mmGroup)
	if err != nil {
		return fieldGroupConversion{}, FunctionNameAndError(n.initFgConversion, err)
	}

	mmGroupConversion := fieldGroupConversion{
		FieldGroups: make([]fieldGroupConversion, 0),
	}

	for _, fgKeySuffix := range mmGroupReadOrderOfFields {
		fgKeySuffixString, err := GetValueAsString(fgKeySuffix)
		if err != nil {
			return fieldGroupConversion{}, FunctionNameAndError(n.initFgConversion, err)
		}

		fgMap, err := GetFieldGroupMap(mmGroupFields[fgKeySuffixString])
		if err != nil {
			return fieldGroupConversion{}, FunctionNameAndError(n.initFgConversion, err)
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

		newFieldGroupConversion := fieldGroupConversion{}
		if fgMapKey, err := GetValueAsString(fgMap[FIELD_GROUP_PROP_FIELD_GROUP_KEY]); err != nil {
			return fieldGroupConversion{}, err
		} else {
			newFieldGroupConversion.FgKey = new(string)
			*newFieldGroupConversion.FgKey = fgMapKey

		}

		if _, err := GetGroupFields(fgMap); err == nil {
			if gReadOrderOfFields, err := GetGroupReadOrderOfFields(fgMap); err == nil {
				if g, err := func() ([]string, error) {
					gReadOrderOfFieldsString := make([]string, len(gReadOrderOfFields))
					for groofIndex, groof := range gReadOrderOfFields {
						if groofString, ok := groof.(string); ok {
							gReadOrderOfFieldsString[groofIndex] = groofString
						} else {
							return nil, argumentsError(n.initFgConversion, "groof", "string", groof)
						}
					}
					return gReadOrderOfFieldsString, nil
				}(); err != nil {
					return fieldGroupConversion{}, FunctionNameAndError(n.initFgConversion, err)
				} else {
					newFieldGroupConversion.GroupReadOrderOfFields = g
				}

				if fgViewMaxNoOfValuesInSeparateColumns := FgGet2DConversion(fgMap); fgViewMaxNoOfValuesInSeparateColumns > 1 {
					newFieldGroupConversion.FgSepColsMaxValues = new(int)
					*newFieldGroupConversion.FgSepColsMaxValues = fgViewMaxNoOfValuesInSeparateColumns
				} else {
					if fgGroupConversion, err := n.initFgConversion(fgMap); err != nil {
						return fieldGroupConversion{}, FunctionNameAndError(n.initFgConversion, err)
					} else {
						newFieldGroupConversion.FieldGroups = fgGroupConversion.FieldGroups
					}
				}

				mmGroupConversion.FieldGroups = append(mmGroupConversion.FieldGroups, newFieldGroupConversion)
				continue
			}
		}

		if fgViewMaxNoOfValuesInSeparateColumns := FgGet2DConversion(fgMap); fgViewMaxNoOfValuesInSeparateColumns > 1 {
			newFieldGroupConversion.FgSepColsMaxValues = new(int)
			*newFieldGroupConversion.FgSepColsMaxValues = fgViewMaxNoOfValuesInSeparateColumns
		}

		if joinSymbol, ok := fgMap[FIELD_GROUP_PROP_FIELD_MULTIPLE_VALUES_JOIN_SYMBOL].(string); ok && len(joinSymbol) > 0 {
			newFieldGroupConversion.FieldJoinSymbol = joinSymbol
		} else {
			newFieldGroupConversion.FieldJoinSymbol = ","
		}

		mmGroupConversion.FieldGroups = append(mmGroupConversion.FieldGroups, newFieldGroupConversion)
	}

	return mmGroupConversion, nil
}

// Converts Data into a 2D array.
//
// Expects data to be a slice of any
func (n *ConvertObjectsTo2DArray) Convert(data any) error {
	if dataSlice, ok := data.([]any); ok {
		if len(dataSlice) < 1 {
			return FunctionNameAndError(n.Convert, errors.New("data is not an slice"))
		}

		for _, datum := range dataSlice {
			if datumCopy, err := deep.Copy(datum); err != nil {
				return err
			} else {
				if datumMap, ok := datumCopy.(map[string]any); ok {
					n.currentDatum = datumMap

					datum2DArray, err := n.convert([][]any{{}}, n.fgConversion, []int{0})
					if err != nil {
						return FunctionNameAndError(n.Convert, err)
					}

					if n.reorder2DFields != nil {
						if err := n.reorder2DFields.Reorder(datum2DArray); err != nil {
							return FunctionNameAndError(n.Convert, err)
						}
					}

					n.array2D = append(n.array2D, datum2DArray...)
					n.currentDatum = make(map[string]any)
				} else {
					return argumentsError(n.Convert, "datum", "map[string]any", datum)
				}
			}
		}
	} else {
		return argumentsError(n.Convert, "data", "[]any", data)
	}

	return nil
}

func (n *ConvertObjectsTo2DArray) convert(datumObject2DArray [][]any, gConversion fieldGroupConversion, arrayIndexes []int) ([][]any, error) {
	if len(gConversion.FieldGroups) < 1 {
		return nil, FunctionNameAndError(n.convert, errors.New("gConversion.FieldGroups is empty"))
	}

	for _, fgConversion := range gConversion.FieldGroups {
		if fgConversion.FgKey == nil {
			return nil, FunctionNameAndError(n.convert, errors.New("fgConversion.FgKey is empty"))
		}

		var valueInObject any
		if pathToValue, err := PreparePathToValueInObject(*fgConversion.FgKey, arrayIndexes); err != nil {
			return nil, FunctionNameAndError(n.convert, err)
		} else {
			valueInObject = intlibjson.GetValueInObject(n.currentDatum, pathToValue)
		}

		if len(fgConversion.GroupReadOrderOfFields) > 0 {
			if fgConversion.FgSepColsMaxValues != nil {
				newValueInObject := make([]any, *fgConversion.FgSepColsMaxValues*len(fgConversion.GroupReadOrderOfFields))

				if valueInObjectSlice, ok := valueInObject.([]any); ok && len(valueInObjectSlice) > 0 {
					startIndexOfValueInObject := 0
					for vioIndex := range valueInObjectSlice {
						if vioIndex == *fgConversion.FgSepColsMaxValues {
							break
						}
						for _, fgKeySuffix := range fgConversion.GroupReadOrderOfFields {
							newValueInObject[startIndexOfValueInObject] = n.getNewFieldValueInObject(intlibjson.GetValueInObject(valueInObject, fmt.Sprintf("$.%d.%s", vioIndex, fgKeySuffix)), fgConversion.FieldJoinSymbol)
							startIndexOfValueInObject += 1
						}
					}
				}

				datumObject2DArray = n.merge2DArrays(datumObject2DArray, [][]any{newValueInObject})
				continue
			}

			if valueInObjectSlice, ok := valueInObject.([]any); ok && len(valueInObjectSlice) > 0 {
				new2DArray := make([][]any, 0)
				for vioIndex := range valueInObjectSlice {
					if value, err := deep.Copy(new2DArray); err != nil {
						return nil, FunctionNameAndError(n.convert, err)
					} else {
						new2DArray = value
					}

					if converted2DArray, err := n.convert(datumObject2DArray, fgConversion, append(arrayIndexes, vioIndex)); err != nil {
						return nil, FunctionNameAndError(n.convert, err)
					} else {
						new2DArray = append(new2DArray, converted2DArray...)
					}
				}
				datumObject2DArray = new2DArray
				continue
			}

			if converted2DArray, err := n.convert(datumObject2DArray, fgConversion, append(arrayIndexes, 0)); err != nil {
				return nil, FunctionNameAndError(n.convert, err)
			} else {
				datumObject2DArray = converted2DArray
				continue
			}
		}

		if fgConversion.FgSepColsMaxValues != nil {
			newValueInObject := make([]any, *fgConversion.FgSepColsMaxValues)

			if valueInObjectSlice, ok := valueInObject.([]any); ok && len(valueInObjectSlice) > 0 {
				for vioIndex, vios := range valueInObjectSlice {
					if vioIndex == *fgConversion.FgSepColsMaxValues {
						break
					}
					newValueInObject[vioIndex] = n.getNewFieldValueInObject(vios, fgConversion.FieldJoinSymbol)
				}
			}

			datumObject2DArray = n.merge2DArrays(datumObject2DArray, [][]any{newValueInObject})
			continue
		}

		datumObject2DArray = n.merge2DArrays(datumObject2DArray, [][]any{{n.getNewFieldValueInObject(valueInObject, fgConversion.FieldJoinSymbol)}})
	}

	return datumObject2DArray, nil
}

func (n *ConvertObjectsTo2DArray) getNewFieldValueInObject(valueInObject any, joinSymbol string) any {
	if valueInObjectSlice, ok := valueInObject.([]any); ok {
		if len(valueInObjectSlice) > 1 {
			newValueInObject := ""
			for vioIndex, vios := range valueInObjectSlice {
				newVios := vios
				if newVios == nil {
					newVios = "null"
				}
				if vioIndex == 0 {
					newValueInObject = fmt.Sprintf("%v", newVios)
				} else {
					newValueInObject = fmt.Sprintf("%s%s%v", newValueInObject, joinSymbol, newVios)
				}
			}
			return newValueInObject
		}

		return valueInObjectSlice[0]
	}

	return valueInObject
}

// Merges rightArray into rightArray.
//
// For example, if size of leftArray is 2 and size of rightArray is 3, the new2Darray will have a size of 6 as each row of leftArray will be merged with each row of rightArray.
func (n *ConvertObjectsTo2DArray) merge2DArrays(leftArray [][]any, rightArray [][]any) [][]any {
	new2DArray := make([][]any, 0)

	for _, lvalue := range leftArray {
		for _, rvalue := range rightArray {
			new2DArray = append(new2DArray, append(lvalue, rvalue...))
		}
	}

	return new2DArray
}

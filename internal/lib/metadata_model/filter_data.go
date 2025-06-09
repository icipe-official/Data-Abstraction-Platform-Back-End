package metadatamodel

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	intlibjson "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/json"
)

// Executes filter conditions against data.
//
// returns an array containing indexes of data that DID NOT pass the filter conditions.
func FilterData(queryConditions []QueryConditions, data []any) ([]int, error) {
	if len(data) == 0 {
		return nil, FunctionNameAndError(FilterData, errors.New("data is empty"))
	}

	filterExcludeDataIndexes := make([]int, 0)

	if len(queryConditions) == 0 {
		return filterExcludeDataIndexes, nil
	}

	for dIndex, dValue := range data {
		filterExcludeDatum := true

		for _, queryCondition := range queryConditions {
			if len(queryCondition) < 1 {
				filterExcludeDatum = false
				break
			}

			queryConditionTrue := true
			for fgKey, fgQueryCondtion := range queryCondition {
				if len(fgQueryCondtion.FilterCondition) < 1 {
					break
				}

				allOrConditionsFalse := true

				for _, orFilterConditions := range fgQueryCondtion.FilterCondition {
					if len(fgQueryCondtion.FilterCondition) < 1 {
						allOrConditionsFalse = false
						break
					}

					allAndConditionsTrue := true
					for _, andCondition := range orFilterConditions {
						if len(fgQueryCondtion.FilterCondition) < 1 {
							break
						}

						if len(andCondition.Condition) < 1 {
							break
						}

						andConditionTrue := false

						var err error
						loopSuccessful := intlibjson.ForEachValueInObject(dValue, fgKey, func(_ []any, valueFound any) bool {
							switch andCondition.Condition {
							case FILTER_CONDTION_NO_OF_ENTRIES_GREATER_THAN, FILTER_CONDTION_NO_OF_ENTRIES_LESS_THAN, FILTER_CONDTION_NO_OF_ENTRIES_EQUAL_TO:
								var valueInt int
								if vInt, ok := andCondition.Value.(int); ok {
									valueInt = vInt
								} else if vFloat, ok := andCondition.Value.(float64); ok {
									valueInt = int(vFloat)
								} else {
									err = argumentsError(FilterData, "filterValue", "int", andCondition.Value)
									return true
								}

								if valueFoundSlice, ok := valueFound.([]any); ok {
									conditionTrue := false
									switch andCondition.Condition {
									case FILTER_CONDTION_NO_OF_ENTRIES_GREATER_THAN:
										conditionTrue = len(valueFoundSlice) > valueInt
									case FILTER_CONDTION_NO_OF_ENTRIES_LESS_THAN:
										conditionTrue = len(valueFoundSlice) < valueInt
									case FILTER_CONDTION_NO_OF_ENTRIES_EQUAL_TO:
										conditionTrue = len(valueFoundSlice) == valueInt
									}
									if conditionTrue {
										if andCondition.Negate {
											return true
										}
										andConditionTrue = true
										return true
									}
								}
							case FILTER_CONDTION_NUMBER_GREATER_THAN, FILTER_CONDTION_NUMBER_LESS_THAN:
								var valueFloat float64
								if vFloat, ok := andCondition.Value.(float64); ok {
									valueFloat = vFloat
								} else if vInt, ok := andCondition.Value.(int); ok {
									valueFloat = float64(vInt)
								} else {
									err = argumentsError(FilterData, "filterValue", "float64", andCondition.Value)
									return true
								}

								if valueFoundSlice, ok := valueFound.([]any); ok {
									conditionTrue := false

									for _, vFound := range valueFoundSlice {
										var valueFoundFloat float64
										if vFloat, ok := vFound.(float64); ok {
											valueFoundFloat = vFloat
										} else if vInt, ok := vFound.(int); ok {
											valueFoundFloat = float64(vInt)
										} else {
											return true
										}

										if isNumberConditionTrue(andCondition.Condition, valueFoundFloat, valueFloat) {
											conditionTrue = true
											break
										}
									}

									if conditionTrue {
										if andCondition.Negate {
											return true
										}
										andConditionTrue = true
										return true
									}
								} else {
									var valueFoundFloat float64
									if vFloat, ok := valueFound.(float64); ok {
										valueFoundFloat = vFloat
									} else if vInt, ok := valueFound.(int); ok {
										valueFoundFloat = float64(vInt)
									} else {
										return true
									}

									if isNumberConditionTrue(andCondition.Condition, valueFoundFloat, valueFloat) {
										if andCondition.Negate {
											return true
										}
										andConditionTrue = true
										return true
									}
								}
							case FILTER_CONDTION_TIMESTAMP_GREATER_THAN, FILTER_CONDTION_TIMESTAMP_LESS_THAN:
								if len(andCondition.DateTimeFormat) == 0 {
									err = argumentsError(FilterData, "andCondition.DateTimeFormat", "string", andCondition.DateTimeFormat)
									return true
								}
								if valueString, ok := andCondition.Value.(string); ok {
									if valueFoundSlice, ok := valueFound.([]any); ok {
										conditionTrue := false

										for _, vFound := range valueFoundSlice {
											if valueFoundString, ok := vFound.(string); ok {
												if isTimestampConditionTrue(andCondition.Condition, andCondition.DateTimeFormat, valueFoundString, valueString) {
													conditionTrue = true
													break
												}
											}

										}

										if conditionTrue {
											if andCondition.Negate {
												return true
											}
											andConditionTrue = true
											return true
										}
									} else {
										if valueFoundString, ok := valueFound.(string); ok {
											if isTimestampConditionTrue(andCondition.Condition, andCondition.DateTimeFormat, valueFoundString, valueString) {
												if andCondition.Negate {
													return true
												}
												andConditionTrue = true
												return true
											}
										}
									}
									break
								} else {
									err = argumentsError(FilterData, "filterValue", "float64", andCondition.Value)
									return true
								}
							case FILTER_CONDTION_TEXT_BEGINS_WITH, FILTER_CONDTION_TEXT_CONTAINS, FILTER_CONDTION_TEXT_ENDS_WITH:
								if valueString, ok := andCondition.Value.(string); ok {
									if valueFoundSlice, ok := valueFound.([]any); ok {
										conditionTrue := false

										for _, vFound := range valueFoundSlice {
											if valueFoundString, ok := vFound.(string); ok {
												if isTextConditionTrue(andCondition.Condition, valueFoundString, valueString) {
													conditionTrue = true
													break
												}
											}

										}

										if conditionTrue {
											if andCondition.Negate {
												return true
											}
											andConditionTrue = true
											return true
										}
									} else {
										if valueFoundString, ok := valueFound.(string); ok {
											if isTextConditionTrue(andCondition.Condition, valueFoundString, valueString) {
												if andCondition.Negate {
													return true
												}
												andConditionTrue = true
												return true
											}
										}
									}
									break
								} else {
									err = argumentsError(FilterData, "filterValue", "float64", andCondition.Value)
									return true
								}
							case FILTER_CONDTION_EQUAL_TO:
								if fEqualToValue := getEqualToValue(andCondition.Value); fEqualToValue != nil {
									if valueFoundSlice, ok := valueFound.([]any); ok {
										conditionTrue := false

										for _, vFound := range valueFoundSlice {
											if isEqualToConditionTrue(fEqualToValue.Type, fEqualToValue.DateTimeFormat, vFound, fEqualToValue.Value) {
												conditionTrue = true
												break
											}
										}

										if conditionTrue {
											if andCondition.Negate {
												return true
											}
											andConditionTrue = true
											return true
										}
									} else {
										if isEqualToConditionTrue(fEqualToValue.Type, fEqualToValue.DateTimeFormat, valueFound, fEqualToValue.Value) {
											if andCondition.Negate {
												return true
											}
											andConditionTrue = true
											return true
										}
									}
								} else {
									err = argumentsError(FilterData, "fEqualToValue", "EqualToFilterValue", andCondition.Value)
									return true
								}
							}

							if andCondition.Negate {
								andConditionTrue = true
								return true
							}

							return false
						})

						if err != nil {
							return nil, err
						}

						if !loopSuccessful && andCondition.Negate {
							andConditionTrue = true
						}

						if !andConditionTrue {
							allAndConditionsTrue = false
							break
						}
					}

					if allAndConditionsTrue {
						allOrConditionsFalse = false
						break
					}
				}

				if allOrConditionsFalse {
					queryConditionTrue = false
					break
				}
			}

			if queryConditionTrue {
				filterExcludeDatum = false
				break
			}
		}

		if filterExcludeDatum {
			filterExcludeDataIndexes = append(filterExcludeDataIndexes, dIndex)
		}
	}

	return filterExcludeDataIndexes, nil
}

type EqualToFilterValue struct {
	Type           string `json:"$TYPE,omitempty"`
	DateTimeFormat string `json:"$DATETIME_FORMAT,omitempty"`
	Value          any    `json:"$VALUE,omitempty"`
}

func getEqualToValue(filterValue any) *EqualToFilterValue {
	if jsonData, err := json.Marshal(filterValue); err != nil {
		return nil
	} else {
		var v *EqualToFilterValue
		if err := json.Unmarshal(jsonData, v); err != nil {
			return nil
		}
		return v
	}
}

func isEqualToConditionTrue(filterValueType string, dateTimeFormat string, valueFound any, filterValue any) bool {
	switch filterValueType {
	case FIELD_SELECT_TYPE_BOOLEAN, FIELD_SELECT_TYPE_TEXT, FIELD_SELECT_TYPE_NUMBER, FIELD_SELECT_TYPE_SELECT:
		return valueFound == filterValue
	case FIELD_SELECT_TYPE_TIMESTAMP:
		if filterValueString, ok := filterValue.(string); ok {
			if valueFoundString, ok := valueFound.(string); ok {
				return isTimestampConditionTrue(FILTER_CONDTION_EQUAL_TO, dateTimeFormat, valueFoundString, filterValueString)
			}
		}
	default:
		return reflect.DeepEqual(valueFound, filterValue)
	}

	return false
}

func isTextConditionTrue(filterCondition string, valueFound string, filterValue string) bool {
	switch filterCondition {
	case FILTER_CONDTION_TEXT_BEGINS_WITH:
		return strings.HasPrefix(valueFound, filterValue)
	case FILTER_CONDTION_TEXT_CONTAINS:
		return strings.HasSuffix(valueFound, filterValue)
	case FILTER_CONDTION_TEXT_ENDS_WITH:
		return strings.Contains(valueFound, filterValue)
	}

	return false
}

func isTimestampConditionTrue(filterCondition string, dateTimeFormat string, valueFound string, filterValue string) bool {
	filterValueDateTime, err := time.Parse(time.RFC3339Nano, filterValue)
	if err != nil {
		return false
	}

	valueFoundDateTime, err := time.Parse(time.RFC3339Nano, valueFound)
	if err != nil {
		return false
	}

	switch filterCondition {
	case FILTER_CONDTION_TIMESTAMP_GREATER_THAN:
		switch dateTimeFormat {
		case FIELD_DATE_TIME_FORMAT_YYYYMMDDHHMM:
			vfYear, vfMonth, vfDay := valueFoundDateTime.Date()
			fvYear, fvMonth, fvDay := valueFoundDateTime.Date()
			if vfYear > fvYear {
				return true
			}
			if vfYear == fvYear {
				if vfMonth > fvMonth {
					return true
				}
				if vfMonth == fvMonth {
					if vfDay > fvDay {
						return true
					}
					if vfDay == fvDay {
						if valueFoundDateTime.Hour() > filterValueDateTime.Hour() {
							return true
						}
						if valueFoundDateTime.Hour() == filterValueDateTime.Hour() {
							if valueFoundDateTime.Minute() > filterValueDateTime.Minute() {
								return true
							}
						}
					}
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYYMMDD:
			vfYear, vfMonth, vfDay := valueFoundDateTime.Date()
			fvYear, fvMonth, fvDay := valueFoundDateTime.Date()
			if vfYear > fvYear {
				return true
			}
			if vfYear == fvYear {
				if vfMonth > fvMonth {
					return true
				}
				if vfMonth == fvMonth {
					if vfDay > fvDay {
						return true
					}
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYYMM:
			vfYear, vfMonth, _ := valueFoundDateTime.Date()
			fvYear, fvMonth, _ := valueFoundDateTime.Date()
			if vfYear > fvYear {
				return true
			}
			if vfYear == fvYear {
				if vfMonth > fvMonth {
					return true
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_HHMM:
			if valueFoundDateTime.Hour() > filterValueDateTime.Hour() {
				return true
			}
			if valueFoundDateTime.Hour() == filterValueDateTime.Hour() {
				if valueFoundDateTime.Minute() > filterValueDateTime.Minute() {
					return true
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYY:
			return valueFoundDateTime.Year() > filterValueDateTime.Year()
		case FIELD_DATE_TIME_FORMAT_MM:
			_, vfMonth, _ := valueFoundDateTime.Date()
			_, fvMonth, _ := valueFoundDateTime.Date()
			return vfMonth > fvMonth
		}
	case FILTER_CONDTION_TIMESTAMP_LESS_THAN:
		switch dateTimeFormat {
		case FIELD_DATE_TIME_FORMAT_YYYYMMDDHHMM:
			vfYear, vfMonth, vfDay := valueFoundDateTime.Date()
			fvYear, fvMonth, fvDay := valueFoundDateTime.Date()
			if vfYear < fvYear {
				return true
			}
			if vfYear == fvYear {
				if vfMonth < fvMonth {
					return true
				}
				if vfMonth == fvMonth {
					if vfDay < fvDay {
						return true
					}
					if vfDay == fvDay {
						if valueFoundDateTime.Hour() < filterValueDateTime.Hour() {
							return true
						}
						if valueFoundDateTime.Hour() == filterValueDateTime.Hour() {
							if valueFoundDateTime.Minute() < filterValueDateTime.Minute() {
								return true
							}
						}
					}
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYYMMDD:
			vfYear, vfMonth, vfDay := valueFoundDateTime.Date()
			fvYear, fvMonth, fvDay := valueFoundDateTime.Date()
			if vfYear < fvYear {
				return true
			}
			if vfYear == fvYear {
				if vfMonth < fvMonth {
					return true
				}
				if vfMonth == fvMonth {
					if vfDay < fvDay {
						return true
					}
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYYMM:
			vfYear, vfMonth, _ := valueFoundDateTime.Date()
			fvYear, fvMonth, _ := valueFoundDateTime.Date()
			if vfYear < fvYear {
				return true
			}
			if vfYear == fvYear {
				if vfMonth < fvMonth {
					return true
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_HHMM:
			if valueFoundDateTime.Hour() < filterValueDateTime.Hour() {
				return true
			}
			if valueFoundDateTime.Hour() == filterValueDateTime.Hour() {
				if valueFoundDateTime.Minute() < filterValueDateTime.Minute() {
					return true
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYY:
			return valueFoundDateTime.Year() < filterValueDateTime.Year()
		case FIELD_DATE_TIME_FORMAT_MM:
			_, vfMonth, _ := valueFoundDateTime.Date()
			_, fvMonth, _ := valueFoundDateTime.Date()
			return vfMonth < fvMonth
		}
	case FILTER_CONDTION_EQUAL_TO:
		switch dateTimeFormat {
		case FIELD_DATE_TIME_FORMAT_YYYYMMDDHHMM:
			vfYear, vfMonth, vfDay := valueFoundDateTime.Date()
			fvYear, fvMonth, fvDay := valueFoundDateTime.Date()
			if vfYear == fvYear {
				if vfMonth == fvMonth {
					if vfDay == fvDay {
						if valueFoundDateTime.Hour() == filterValueDateTime.Hour() {
							if valueFoundDateTime.Minute() == filterValueDateTime.Minute() {
								return true
							}
						}
					}
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYYMMDD:
			vfYear, vfMonth, vfDay := valueFoundDateTime.Date()
			fvYear, fvMonth, fvDay := valueFoundDateTime.Date()
			if vfYear == fvYear {
				if vfMonth == fvMonth {
					if vfDay == fvDay {
						return true
					}
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYYMM:
			vfYear, vfMonth, _ := valueFoundDateTime.Date()
			fvYear, fvMonth, _ := valueFoundDateTime.Date()
			if vfYear == fvYear {
				if vfMonth == fvMonth {
					return true
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_HHMM:
			if valueFoundDateTime.Hour() == filterValueDateTime.Hour() {
				if valueFoundDateTime.Minute() == filterValueDateTime.Minute() {
					return true
				}
			}
			return false
		case FIELD_DATE_TIME_FORMAT_YYYY:
			return valueFoundDateTime.Year() == filterValueDateTime.Year()
		case FIELD_DATE_TIME_FORMAT_MM:
			_, vfMonth, _ := valueFoundDateTime.Date()
			_, fvMonth, _ := valueFoundDateTime.Date()
			return vfMonth == fvMonth
		}
	}

	return false
}

func isNumberConditionTrue(filterCondtion string, valueFound float64, filterValue float64) bool {
	switch filterCondtion {
	case FILTER_CONDTION_NUMBER_GREATER_THAN:
		return valueFound > filterValue
	case FILTER_CONDTION_NO_OF_ENTRIES_LESS_THAN:
		return valueFound < filterValue
	}

	return false
}

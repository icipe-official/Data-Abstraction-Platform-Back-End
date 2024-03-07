package lib

import (
	"data_administration_platform/internal/pkg/lib"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	jet "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func SendFile(w http.ResponseWriter, fileContentType *string, file *os.File) {
	const currentSection = "File Stream"
	var fileBuffer = make([]byte, CHUNK_SIZE)
	w.Header().Set("Content-Type", *fileContentType)
	w.Header().Set("Cache-Control", "private, max-age=0")
	w.WriteHeader(http.StatusOK)
	for i := 0; ; i++ {
		if noOfBytes, err := file.Read(fileBuffer); err != nil {
			if err != io.EOF {
				lib.Log(lib.LOG_ERROR, currentSection, fmt.Sprintf("Read file in chunks failed | reason: %v", err))
			}
			break
		} else {
			lib.Log(lib.LOG_DEBUG, currentSection, fmt.Sprintf("%v: Reading %v bytes from file", i+1, noOfBytes))
			w.Write(fileBuffer[:noOfBytes])
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			} else {
				lib.Log(lib.LOG_ERROR, currentSection, fmt.Sprintf("Read file in chunks failed | reason: %v", "Could not create flusher"))
			}
		}
	}
}

func genColumnHeaders(columnHeaders *[]string, currentTemplateSection map[string]interface{}) {
	currentColumnHeaders := currentTemplateSection["columns"]
	if fmt.Sprintf("%T", currentColumnHeaders) == "[]interface {}" {
		var currentTemplateSectionValue any
		if fmt.Sprintf("%T", currentTemplateSection["value"]) == "[]interface {}" {
			currentTemplateSectionValue = currentTemplateSection["value"].([]interface{})[0]
		} else {
			currentTemplateSectionValue = currentTemplateSection["value"]
		}
		if fmt.Sprintf("%T", currentTemplateSectionValue) == "map[string]interface {}" {
			for _, c := range currentColumnHeaders.([]interface{}) {
				if strings.HasPrefix(c.(string), "@") {
					if regexp.MustCompile(`renderhorizontally=true&#`).MatchString(currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["struct"].(string)) &&
						fmt.Sprintf("%T", currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["value"]) == "[]interface {}" {
						maxNoOfResultsExtract := regexp.MustCompile(`max=(.+?)&#`).FindStringSubmatch(currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["struct"].(string))
						if len(maxNoOfResultsExtract) == 2 {
							if max, err := strconv.Atoi(maxNoOfResultsExtract[1]); err == nil {
								columns := []string{}
								for _, key := range currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["columns"].([]interface{}) {
									value := currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["value"].([]interface{})[0].(map[string]interface{})[key.(string)].(map[string]interface{})
									fieldName := ""
									fieldNameExtract := regexp.MustCompile(`name=(.+?)&#`).FindStringSubmatch(value["struct"].(string))
									if len(fieldNameExtract) == 2 {
										fieldName = fieldNameExtract[1]
									} else {
										fieldNames := strings.Split(strings.Split(value["struct"].(string), " ")[0], ".")
										fieldName = fieldNames[len(fieldNames)-1]
									}
									columns = append(columns, fieldName)
								}
								for i := 0; i < max; i++ {
									for _, c := range columns {
										*columnHeaders = append((*columnHeaders), fmt.Sprintf("%v_%v", c, i+1))
									}
								}
							} else {
								genColumnHeaders(columnHeaders, currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{}))
							}
						} else {
							genColumnHeaders(columnHeaders, currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{}))
						}
					} else {
						genColumnHeaders(columnHeaders, currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{}))
					}
				} else {
					if field, ok := currentTemplateSectionValue.(map[string]interface{})[c.(string)]; ok && fmt.Sprintf("%T", field) == "map[string]interface {}" {
						if fieldstruct, ok := field.(map[string]interface{})["struct"]; ok && fmt.Sprintf("%T", fieldstruct) == "string" {
							fieldNameExtract := regexp.MustCompile(`name=(.+?)&#`).FindStringSubmatch(fieldstruct.(string))
							columnName := ""
							if len(fieldNameExtract) == 2 {
								columnName = fieldNameExtract[1]
							} else {
								columnName = c.(string)
							}
							if regexp.MustCompile(`renderhorizontally=true&#`).MatchString(fieldstruct.(string)) {
								maxNoOfResultsExtract := regexp.MustCompile(`max=(.+?)&#`).FindStringSubmatch(fieldstruct.(string))
								if len(maxNoOfResultsExtract) == 2 {
									if max, err := strconv.Atoi(maxNoOfResultsExtract[1]); err == nil {
										for i := 0; i < max; i++ {
											*columnHeaders = append((*columnHeaders), fmt.Sprintf("%v_%v", columnName, i+1))
										}
									} else {
										*columnHeaders = append((*columnHeaders), columnName)
									}
								}
							} else {
								*columnHeaders = append((*columnHeaders), columnName)
							}

						}
					}
				}
			}
		}
	}
}
func GetColumnHeadersForTwoDimensionArray(template map[string]interface{}) []string {
	columnHeaders := make([]string, 0)
	genColumnHeaders(&columnHeaders, template)
	return columnHeaders
}

func mergeTwoDimensionArrays(leftArray, rightArray [][]any) [][]any {
	newArray := [][]any{}
	for _, l := range leftArray {
		for _, r := range rightArray {
			newArray = append(newArray, append(l, r...))
		}
	}
	return newArray
}

func ConvertMapIntoTwoDimensionArray(twoDimensionArray [][]any, currentTemplateSection, value map[string]interface{}, repetitiveIndexes []int) [][]any {
	columns := currentTemplateSection["columns"]
	if fmt.Sprintf("%T", columns) == "[]interface {}" {
		var currentTemplateSectionValue any
		if fmt.Sprintf("%T", currentTemplateSection["value"]) == "[]interface {}" {
			currentTemplateSectionValue = currentTemplateSection["value"].([]interface{})[0]
		} else {
			currentTemplateSectionValue = currentTemplateSection["value"]
		}
		if fmt.Sprintf("%T", currentTemplateSectionValue) == "map[string]interface {}" {
			for _, c := range columns.([]interface{}) {
				if strings.HasPrefix(c.(string), "@") {
					if group, ok := currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)]; ok && fmt.Sprintf("%T", group) == "map[string]interface {}" {
						if groupStruct, ok := group.(map[string]interface{})["struct"]; ok && fmt.Sprintf("%T", groupStruct) == "string" {
							if strings.Split(groupStruct.(string), " ")[1] == "unique" {
								twoDimensionArray = ConvertMapIntoTwoDimensionArray(twoDimensionArray, group.(map[string]interface{}), value, repetitiveIndexes)
							} else {
								if keyPathInValue, err := prepareKey(strings.ReplaceAll(strings.Replace(strings.Split(groupStruct.(string), " ")[0], "root.", "", 1), ".value", ""), repetitiveIndexes); err != nil {
									return [][]any{}
								} else {
									groupValue := GetValueInMap(value, keyPathInValue)
									if regexp.MustCompile(`renderhorizontally=true&#`).MatchString(currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["struct"].(string)) &&
										fmt.Sprintf("%T", currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["value"]) == "[]interface {}" {
										maxNoOfResultsExtract := regexp.MustCompile(`max=(.+?)&#`).FindStringSubmatch(currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["struct"].(string))
										if len(maxNoOfResultsExtract) == 2 {
											if max, err := strconv.Atoi(maxNoOfResultsExtract[1]); err == nil && max > 0 {
												newSingleDimensionArray := []any{}
												if fmt.Sprintf("%T", groupValue) == "[]interface {}" {
													for i := 0; i < max; i++ {
														if i < len(groupValue.([]interface{})) {
															for _, key := range group.(map[string]interface{})["columns"].([]interface{}) {
																if groupValue.([]interface{})[i] != nil && fmt.Sprintf("%T", groupValue.([]interface{})[i].(map[string]interface{})[key.(string)]) == "map[string]interface {}" {
																	newSingleDimensionArray = append(newSingleDimensionArray, "#SECTION HAS NESTED GROUPS")
																} else if groupValue.([]interface{})[i] != nil && fmt.Sprintf("%T", groupValue.([]interface{})[i].(map[string]interface{})[key.(string)]) != "map[string]interface {}" {
																	newSingleDimensionArray = append(newSingleDimensionArray, groupValue.([]interface{})[i].(map[string]interface{})[key.(string)])
																} else {
																	newSingleDimensionArray = append(newSingleDimensionArray, nil)
																}
															}
														} else {
															for range group.(map[string]interface{})["columns"].([]interface{}) {
																newSingleDimensionArray = append(newSingleDimensionArray, nil)
															}
														}
													}
												} else {
													for i := 0; i < max; i++ {
														for range group.(map[string]interface{})["columns"].([]interface{}) {
															newSingleDimensionArray = append(newSingleDimensionArray, nil)
														}
													}
												}
												twoDimensionArray = mergeTwoDimensionArrays(twoDimensionArray, [][]any{newSingleDimensionArray})
											} else {
												if fmt.Sprintf("%T", groupValue) == "[]interface {}" {
													newDimensionArray := [][]any{}
													for i := range groupValue.([]interface{}) {
														leftDimensionArray := make([][]any, len(newDimensionArray))
														for i, v := range newDimensionArray {
															leftDimensionArray[i] = make([]any, len(v))
															copy(leftDimensionArray[i], v)
														}
														newDimensionArray = slices.Concat(leftDimensionArray, ConvertMapIntoTwoDimensionArray(twoDimensionArray, group.(map[string]interface{}), value, append(repetitiveIndexes, i)))
														if strings.Contains(keyPathInValue, "ir_bioassay") {
															fmt.Printf("%v -> %v -> Right Dimension: %v\n", append(repetitiveIndexes, i), len(newDimensionArray), newDimensionArray)
														}
													}
													twoDimensionArray = newDimensionArray
												} else {
													twoDimensionArray = ConvertMapIntoTwoDimensionArray(twoDimensionArray, group.(map[string]interface{}), value, append(repetitiveIndexes, 0))
												}
											}
										} else {
											if fmt.Sprintf("%T", groupValue) == "[]interface {}" {
												newDimensionArray := [][]any{}
												for i := range groupValue.([]interface{}) {
													leftDimensionArray := make([][]any, len(newDimensionArray))
													for i, v := range newDimensionArray {
														leftDimensionArray[i] = make([]any, len(v))
														copy(leftDimensionArray[i], v)
													}
													newDimensionArray = slices.Concat(leftDimensionArray, ConvertMapIntoTwoDimensionArray(twoDimensionArray, group.(map[string]interface{}), value, append(repetitiveIndexes, i)))
													if strings.Contains(keyPathInValue, "ir_bioassay") {
														fmt.Printf("%v -> %v -> Right Dimension: %v\n", append(repetitiveIndexes, i), len(newDimensionArray), newDimensionArray)
													}
												}
												twoDimensionArray = newDimensionArray
											} else {
												twoDimensionArray = ConvertMapIntoTwoDimensionArray(twoDimensionArray, group.(map[string]interface{}), value, append(repetitiveIndexes, 0))
											}
										}
									} else {
										if fmt.Sprintf("%T", groupValue) == "[]interface {}" {
											newDimensionArray := [][]any{}
											for i := range groupValue.([]interface{}) {
												leftDimensionArray := make([][]any, len(newDimensionArray))
												for i, v := range newDimensionArray {
													leftDimensionArray[i] = make([]any, len(v))
													copy(leftDimensionArray[i], v)
												}
												newDimensionArray = slices.Concat(leftDimensionArray, ConvertMapIntoTwoDimensionArray(twoDimensionArray, group.(map[string]interface{}), value, append(repetitiveIndexes, i)))
												if strings.Contains(keyPathInValue, "ir_bioassay") {
													fmt.Printf("%v -> %v -> Right Dimension: %v\n", append(repetitiveIndexes, i), len(newDimensionArray), newDimensionArray)
												}
											}
											twoDimensionArray = newDimensionArray
										} else {
											twoDimensionArray = ConvertMapIntoTwoDimensionArray(twoDimensionArray, group.(map[string]interface{}), value, append(repetitiveIndexes, 0))
										}
									}
								}
							}
						}
					}

				} else {
					if field, ok := currentTemplateSectionValue.(map[string]interface{})[c.(string)]; ok && fmt.Sprintf("%T", field) == "map[string]interface {}" {
						if fieldstruct, ok := field.(map[string]interface{})["struct"]; ok && fmt.Sprintf("%T", fieldstruct) == "string" {
							if keyPathInValue, err := prepareKey(strings.ReplaceAll(strings.Replace(strings.Split(fieldstruct.(string), " ")[0], "root.", "", 1), ".value", ""), repetitiveIndexes); err != nil {
								return [][]any{}
							} else {
								fieldValue := GetValueInMap(value, keyPathInValue)
								if regexp.MustCompile(`renderhorizontally=true&#`).MatchString(fieldstruct.(string)) {
									maxNoOfResultsExtract := regexp.MustCompile(`max=(.+?)&#`).FindStringSubmatch(currentTemplateSectionValue.(map[string]interface{})[strings.Replace(c.(string), "@", "", 1)].(map[string]interface{})["struct"].(string))
									if len(maxNoOfResultsExtract) == 2 {
										if max, err := strconv.Atoi(maxNoOfResultsExtract[1]); err == nil {
											newSingleDimensionArray := []any{}
											if fmt.Sprintf("%T", fieldValue) == "[]interface {}" {
												for i := 0; i < max; i++ {
													if i >= len(fieldValue.([]interface{})) {
														newSingleDimensionArray = append(newSingleDimensionArray, nil)
													} else {
														newSingleDimensionArray = append(newSingleDimensionArray, fieldValue.([]interface{})[i])
													}
												}
											} else {
												for i := 0; i < max; i++ {
													newSingleDimensionArray = append(newSingleDimensionArray, nil)
												}
											}
											twoDimensionArray = mergeTwoDimensionArrays(twoDimensionArray, [][]any{newSingleDimensionArray})
										} else {
											if fmt.Sprintf("%T", fieldValue) == "[]interface {}" {
												fieldValue = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(fieldValue.([]interface{}))), ", "), "[]")
											}
											twoDimensionArray = mergeTwoDimensionArrays(twoDimensionArray, [][]any{{fieldValue}})
										}
									} else {
										if fmt.Sprintf("%T", fieldValue) == "[]interface {}" {
											fieldValue = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(fieldValue.([]interface{}))), ", "), "[]")
										}
										twoDimensionArray = mergeTwoDimensionArrays(twoDimensionArray, [][]any{{fieldValue}})
									}
								} else {
									if fmt.Sprintf("%T", fieldValue) == "[]interface {}" {
										fieldValue = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(fieldValue.([]interface{}))), ", "), "[]")
									}
									twoDimensionArray = mergeTwoDimensionArrays(twoDimensionArray, [][]any{{fieldValue}})
								}
							}
						}
					}
				}
			}
		}
	}
	return twoDimensionArray
}

// returns error if key still contains '[x]' after processing repetitiveIndexes
func prepareKey(key string, repetitiveIndexes []int) (string, error) {
	for _, ri := range repetitiveIndexes {
		key = strings.Replace(key, "[x]", fmt.Sprintf("[%v]", ri), 1)
	}

	if strings.Contains(key, "[x]") {
		return "", fmt.Errorf("key %v does not match indexes %v", key, repetitiveIndexes)
	}

	key = strings.Replace(key, "root.", "", 1)
	return key, nil
}

// currentValue expected to be a map of string keys with any values
func SetValueInMap(currentValue any, path string, valueToSet any) any {
	if currentValue == nil {
		currentValue = map[string]interface{}{}
	}
	paths := strings.Split(path, ".")
	pathAndIndex := regexp.MustCompile(`(.+?)\[(\d+)\]`).FindStringSubmatch(paths[0])
	if len(paths) == 1 {
		if len(pathAndIndex) == 3 {
			if arrayIndex, err := strconv.Atoi(pathAndIndex[2]); err == nil && arrayIndex >= 0 {
				if val, ok := currentValue.(map[string]interface{})[pathAndIndex[1]]; !ok || fmt.Sprintf("%T", val) != "[]interface {}" {
					currentValue.(map[string]interface{})[pathAndIndex[1]] = []interface{}{}
				}
				if arrayIndex >= len(currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{})) {
					for i := len(currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{})); i <= arrayIndex; i++ {
						currentValue.(map[string]interface{})[pathAndIndex[1]] = append(currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{}), nil)
					}
				}
				currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{})[arrayIndex] = valueToSet
			}
		} else {
			currentValue.(map[string]interface{})[paths[0]] = valueToSet
		}
	} else {
		if len(pathAndIndex) == 3 {
			if arrayIndex, err := strconv.Atoi(pathAndIndex[2]); err == nil && arrayIndex >= 0 {
				if val, ok := currentValue.(map[string]interface{})[pathAndIndex[1]]; !ok || fmt.Sprintf("%T", val) != "[]interface {}" {
					currentValue.(map[string]interface{})[pathAndIndex[1]] = []interface{}{}
				}
				if arrayIndex >= len(currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{})) {
					for i := len(currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{})); i <= arrayIndex; i++ {
						currentValue.(map[string]interface{})[pathAndIndex[1]] = append(currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{}), nil)
					}
				}
				currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{})[arrayIndex] = SetValueInMap(currentValue.(map[string]interface{})[pathAndIndex[1]].([]interface{})[arrayIndex], strings.Join(paths[1:], "."), valueToSet)
			}
		} else {
			if val, ok := currentValue.(map[string]interface{})[paths[0]]; !ok || fmt.Sprintf("%T", val) != "map[string]interface {}" {
				currentValue.(map[string]interface{})[paths[0]] = map[string]interface{}{}
			}
			currentValue.(map[string]interface{})[paths[0]] = SetValueInMap(currentValue.(map[string]interface{})[paths[0]], strings.Join(paths[1:], "."), valueToSet)
		}
	}

	return currentValue
}

// currentValue expected to be a map of string keys with any values
func GetValueInMap(currentValue any, path string) any {
	if fmt.Sprintf("%T", currentValue) == "map[string]interface {}" {
		paths := strings.Split(path, ".")
		pathAndIndex := regexp.MustCompile(`(.+?)\[(\d+)\]`).FindStringSubmatch(paths[0])
		if len(paths) == 1 {
			if len(pathAndIndex) == 3 {
				if arrayIndex, err := strconv.Atoi(pathAndIndex[2]); err == nil && arrayIndex >= 0 {
					if val, ok := currentValue.(map[string]interface{})[pathAndIndex[1]]; ok && fmt.Sprintf("%T", val) == "[]interface {}" {
						if arrayIndex < len(val.([]interface{})) {
							return val.([]interface{})[arrayIndex]
						}
					}
				}
			} else {
				if val, ok := currentValue.(map[string]interface{})[paths[0]]; ok {
					return val
				}
			}
		} else {
			if len(pathAndIndex) == 3 {
				if arrayIndex, err := strconv.Atoi(pathAndIndex[2]); err == nil && arrayIndex >= 0 {
					if val, ok := currentValue.(map[string]interface{})[pathAndIndex[1]]; ok && fmt.Sprintf("%T", val) == "[]interface {}" {
						if arrayIndex < len(val.([]interface{})) {
							return GetValueInMap(val.([]interface{})[arrayIndex], strings.Join(paths[1:], "."))
						}
					}
				}
			} else if val, ok := currentValue.(map[string]any)[paths[0]]; ok {
				return GetValueInMap(val, strings.Join(paths[1:], "."))
			}
		}
	}

	return nil
}

func ValidateEmailEnv(currentSection string) bool {
	envVariablesRequired := []string{
		"MAIL_HOST",
		"MAIL_PORT",
		"MAIL_USERNAME",
		"MAIL_PASSWORD",
	}
	isEnvValid := true
	for _, evr := range envVariablesRequired {
		if os.Getenv(evr) == "" {
			lib.Log(lib.LOG_ERROR, currentSection, fmt.Sprintf("%v not set", evr))
			isEnvValid = false
		}
	}
	return isEnvValid
}

func GenRandomString(length int) string {
	stringPool := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_~-")
	randomString := make([]rune, length)
	for index := range randomString {
		randomString[index] = stringPool[rand.Intn(len(stringPool))]
	}
	return string(randomString)
}

func CtxGetCurrentUser(r *http.Request) User {
	return r.Context().Value(CURRENT_USER_CTX_KEY).(User)
}

func GetUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func NewError(code int, message string) error {
	return fmt.Errorf("%v%v%v", code, lib.OPTS_SPLIT, message)
}

func GetTextSearchBoolExp(vectorColumn string, searchQuery string) jet.Expression {
	sqSplitSpace := strings.Split(searchQuery, " ")
	if len(sqSplitSpace) > 0 {
		newQuery := fmt.Sprintf("%v @@ to_tsquery('%v:*')", vectorColumn, sqSplitSpace[0])
		for i := 1; i < len(sqSplitSpace); i++ {
			newQuery = newQuery + " AND " + fmt.Sprintf("%v @@ to_tsquery('%v:*')", vectorColumn, sqSplitSpace[i])
		}
		return jet.Raw(newQuery)
	} else {
		return jet.Raw(fmt.Sprintf("%v @@ to_tsquery('%v:*')", vectorColumn, searchQuery))
	}
}

func SendJsonResponse[T any](data T, w http.ResponseWriter) {
	response, err := prepareJsonResponse(data, w)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func SendErrorResponse(err error, w http.ResponseWriter) {
	newError := strings.Split(err.Error(), lib.OPTS_SPLIT)
	statusCode, err2 := strconv.Atoi(newError[0])
	if err2 != nil {
		lib.Log(lib.LOG_ERROR, "Utils", fmt.Sprintf("Error preparing error response | reason: %v", err2))
		errorMessage := JsonMessage{Message: "Internal Server Error"}
		response, err2 := prepareJsonResponse(errorMessage, w)
		if err2 != nil {
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(response)
		return
	}
	errorMessage := JsonMessage{Message: newError[1]}
	response, err2 := prepareJsonResponse(errorMessage, w)
	if err2 != nil {
		return
	}
	w.WriteHeader(statusCode)
	w.Write(response)
}

func prepareJsonResponse[T any](value T, w http.ResponseWriter) ([]byte, error) {
	json, err := json.Marshal(value)
	if err != nil {
		errorMessage := JsonMessage{Message: "Internal Server Error"}
		log.Printf("Error preparing json error message | reason: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errorMessage.Message))
		return nil, err
	}
	w.Header().Set("Content-Type", "application/json")
	return json, nil
}

func getRedisSessionDb() int {
	if rsd, err := strconv.Atoi(os.Getenv("REDIS_SESSION_DB")); err != nil {
		return 15
	} else {
		return rsd
	}
}

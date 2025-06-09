package metadatamodel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/brunoga/deep"
	intlibjson "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/json"
)

func TestPreparePathToValueInObject(t *testing.T) {
	testData := []struct {
		path         string
		groupIndexes []int
		expectedPath string
	}{
		{path: "$.$GROUP_FIELDS[*].child.$GROUP_FIELDS[*].nectar.$GROUP_FIELDS[*].willy.$GROUP_FIELDS[*].oxford", groupIndexes: []int{0, 1, 0, 15}, expectedPath: "$.child[1].nectar[0].willy[15].oxford"},
		{path: "$.$GROUP_FIELDS[*].child.$GROUP_FIELDS[*].nectar.$GROUP_FIELDS[*].willy", groupIndexes: []int{0, 5, 3}, expectedPath: "$.child[5].nectar[3].willy"},
		{path: "$.$GROUP_FIELDS[*].child.$GROUP_FIELDS[*].nectar.$GROUP_FIELDS[*].willy.$GROUP_FIELDS[*].wonka", groupIndexes: []int{0, 5, 3}},
		{path: "$GROUP_FIELDS[*].child.$GROUP_FIELDS[*]", groupIndexes: []int{4}},
		{path: "$GROUP_FIELDS.[*].child.$GROUP_FIELDS.[*].walker", groupIndexes: []int{0, 1}},
	}

	for _, value := range testData {
		expectedPath, err := PreparePathToValueInObject(value.path, value.groupIndexes)
		if len(value.expectedPath) > 0 {
			if err != nil {
				t.Errorf("\nFAILED, error in processing path, error: %v\nreturnedPath:%+v\ntestData:%+v", err, expectedPath, value)
			} else {
				if value.expectedPath != expectedPath {
					t.Errorf("\nFAILED, expected path to be %v, found %v\ntestData:%+v", value.expectedPath, expectedPath, value)
				}
			}
		} else {
			if err == nil {
				t.Errorf("\nFAILED, expected error in processing path, found %v\ntestData:%+v", expectedPath, value)
			}
		}
	}
}

func TestMetadataModelAndDataManipulation(t *testing.T) {
	testDataDirectory, err := filepath.Abs("../../../test_data")
	if err != nil {
		t.Fatalf("\nERROR, could not get location where test_data is store, err: %v", err)
	}

	type td struct {
		file string
		data any
	}
	testData := map[string]td{
		"test_data_2darray":  {file: "test_data_2darray.json"},
		"test_data":          {file: "test_data.json"},
		"test_metadatamodel": {file: "test_metadatamodel.json"},
	}

	for key, value := range testData {
		if data, err := os.ReadFile(fmt.Sprintf("%v/%v", testDataDirectory, value.file)); err != nil {
			t.Fatalf("\nERROR, could not open %v, err: %v", value.file, err)
		} else {
			var parsedData any

			if err := json.Unmarshal(data, &parsedData); err != nil {
				t.Fatalf("\nERROR, could not parse contents of file %v as json, err: %v", value.file, err)
			} else {
				testData[key] = td{
					file: value.file,
					data: parsedData,
				}
			}
		}
	}

	objectTo2DArray, err := NewConvertObjectsTo2DArray(testData["test_metadatamodel"].data, nil, true, true)
	if err != nil {
		t.Fatalf("\nERROR, could not init objectTo2DArray, err: %v", err)
	}
	if err := objectTo2DArray.Convert(testData["test_data"].data); err != nil {
		t.Fatalf("\nERROR, execute objectTo2DArray.Convert, err: %v", err)
	}
	if len(testData["test_data_2darray"].data.([]any)) != len(objectTo2DArray.Array2D()) {
		t.Fatalf("\nERROR, length of objectTo2DArray.Array2D() [%d] and testData[\"test_data_2darray\"].data [%d] not equal", len(objectTo2DArray.Array2D()), len(testData["test_data_2darray"].data.([]any)))
	}
	someRowsNotEqual := false
	for index, value := range testData["test_data_2darray"].data.([]any) {
		if !reflect.DeepEqual(objectTo2DArray.Array2D()[index], testData["test_data_2darray"].data.([]any)[index]) {
			someRowsNotEqual = true
			fmt.Printf("%d original (%d) |-> %+v\n\n", index, len(value.([]any)), value.([]any))
			fmt.Printf("%d converted (%d) |-> %+v\n\n", index, len(objectTo2DArray.Array2D()[index]), objectTo2DArray.Array2D()[index])
		}
	}
	if someRowsNotEqual {
		t.Fatalf("\nERROR, objectTo2DArray.Array2D() and testData[\"test_data_2darray\"].data not equal")
	}

	array2DToObject, err := NewConvert2DArrayToObjects(testData["test_metadatamodel"].data, nil, true, true, nil)
	if err != nil {
		t.Fatalf("\nERROR, could not init arrray2DToObject, err: %v", err)
	}
	if err := array2DToObject.Convert(testData["test_data_2darray"].data); err != nil {
		t.Fatalf("\nERROR, execute objectTo2DArray.Convert, err: %v", err)
	}
	if len(testData["test_data"].data.([]any)) != len(array2DToObject.Objects()) {
		fmt.Println()
		t.Fatalf("\nERROR, length of arrray2DToObject.Objects() [%d] and testData[\"test_data\"].data [%d] not equal", len(array2DToObject.Objects()), len(testData["test_data"].data.([]any)))
	}
	someRowsNotEqual = false
	for index, value := range testData["test_data"].data.([]any) {
		if !intlibjson.AreValuesEqual(array2DToObject.Objects()[index], testData["test_data"].data.([]any)[index]) {
			someRowsNotEqual = true
			fmt.Printf("%d original |-> %+v\n\n", index, value)
			fmt.Printf("%d converted |-> %+v\n\n", index, array2DToObject.Objects()[index])
		}
	}
	if someRowsNotEqual {
		t.Fatalf("\nERROR, arrray2DToObject.Objects() and testData[\"test_data\"].data not equal")
	}

	objectTo2DArrayClone, err := deep.Copy(objectTo2DArray.Array2D())
	if err != nil {
		t.Fatalf("\nERROR, deep copy objectTo2DArrayClone failed, error: %v", err)
	}
	array2DToObject.ResetObjects()
	if err := array2DToObject.Convert(objectTo2DArrayClone); err != nil {
		t.Fatalf("\nERROR, execute array2DToObject.Convert(objectTo2DArrayClone) failed, error: %v", err)
	}

	someRowsNotEqual = false
	for index, value := range testData["test_data"].data.([]any) {
		if !intlibjson.AreValuesEqual(array2DToObject.Objects()[index], testData["test_data"].data.([]any)[index]) {
			someRowsNotEqual = true
			fmt.Printf("%d original |-> %+v\n\n", index, value)
			fmt.Printf("%d converted |-> %+v\n\n", index, array2DToObject.Objects()[index])
		}
	}
	if someRowsNotEqual {
		t.Fatalf("\nERROR, Back and forth conversion arrray2DToObject.Objects() and testData[\"test_data\"].data not equal")
	}

	array2DToObjectClone, err := deep.Copy(array2DToObject.Objects())
	if err != nil {
		t.Fatalf("\nERROR, deep copy array2DToObjectClone failed, error: %v", err)
	}
	objectTo2DArray.ResetArray2D()
	if err := objectTo2DArray.Convert(array2DToObjectClone); err != nil {
		t.Fatalf("\nERROR, execute objectTo2DArray.Convert(array2DToObjectClone) failed, error: %v", err)
	}
	someRowsNotEqual = false
	for index, value := range testData["test_data_2darray"].data.([]any) {
		if !reflect.DeepEqual(objectTo2DArray.Array2D()[index], testData["test_data_2darray"].data.([]any)[index]) {
			someRowsNotEqual = true
			fmt.Printf("%d original (%d) |-> %+v\n\n", index, len(value.([]any)), value.([]any))
			fmt.Printf("%d converted (%d) |-> %+v\n\n", index, len(objectTo2DArray.Array2D()[index]), objectTo2DArray.Array2D()[index])
		}
	}
	if someRowsNotEqual {
		t.Fatalf("\nERROR, Back and forth objectTo2DArray.Array2D() and testData[\"test_data_2darray\"].data not equal")
	}
}

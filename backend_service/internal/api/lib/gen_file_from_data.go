package lib

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type FileFromData struct {
	SingleSheet     string `json:"-"`
	FileType        string `json:"-"`
	Data            []map[string]interface{}
	ModelTemplate   map[string]interface{} `json:"-"`
	FilePath        string
	FileContentType string
}

func (n *FileFromData) GenFileFromData() error {
	switch n.FileType {
	case GEN_FILE_CSV:
		n.FileContentType = "text/csv"
		n.FilePath = fmt.Sprintf("%v/%v_%v.csv", TMP_DIR, time.Now().Unix(), uuid.New())
		newFile, err := os.Create(n.FilePath)
		if err != nil {
			return fmt.Errorf("create new csv file failed | reason: %v", err)
		}
		defer newFile.Close()

		csvWriter := csv.NewWriter(newFile)
		defer csvWriter.Flush()

		csvWriter.Write(GetColumnHeadersForTwoDimensionArray(n.ModelTemplate))
		func() {
			for i, d := range n.Data {
				for _, row := range ConvertMapIntoTwoDimensionArray([][]any{{}}, n.ModelTemplate, d, []int{}) {
					newRow := []string{}
					for _, rd := range row {
						if rd != nil {
							newRow = append(newRow, fmt.Sprintf("%v", rd))
						} else {
							newRow = append(newRow, "")
						}
					}
					csvWriter.Write(newRow)
				}
				n.Data[i] = map[string]interface{}{}
			}
		}()
	case GEN_FILE_EXCEL:
		newWorkBook := excelize.NewFile()
		defer newWorkBook.Close()

		streamWriter, err := newWorkBook.NewStreamWriter("Sheet1")
		if err != nil {
			return fmt.Errorf("create stream writer for workbook failed | reason: %v", err)
		}

		if n.SingleSheet == "true" {
			if err := func() error {
				excelColumnHeaders := []interface{}{}
				for _, ch := range GetColumnHeadersForTwoDimensionArray(n.ModelTemplate) {
					excelColumnHeaders = append(excelColumnHeaders, excelize.Cell{Value: ch})
				}
				if err := streamWriter.SetRow(fmt.Sprintf("A%v", 1), excelColumnHeaders); err != nil {
					return fmt.Errorf("set column headers in excel sheet of new workbook failed | reason: %v", err)
				}
				return nil
			}(); err != nil {
				return err
			} else {
				if err := func() error {
					currentRow := 1
					for i, d := range n.Data {
						for _, row := range ConvertMapIntoTwoDimensionArray([][]any{{}}, n.ModelTemplate, d, []int{}) {
							currentRow += 1
							if err := streamWriter.SetRow(fmt.Sprintf("A%v", currentRow), row); err != nil {
								return fmt.Errorf("set column headers in excel sheet of new workbook failed | reason: %v", err)
							}
						}
						n.Data[i] = map[string]interface{}{}
					}
					return nil
				}(); err != nil {
					return err
				}
			}
		}
		if err := streamWriter.Flush(); err != nil {
			return fmt.Errorf("save excel buffer content failed | reason: %v", err)
		}
		n.FileContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		n.FilePath = fmt.Sprintf("%v/%v_%v.xlsx", TMP_DIR, time.Now().UTC(), uuid.New())
		if err := newWorkBook.SaveAs(n.FilePath); err != nil {
			return fmt.Errorf("save new workbook failed | reason: %v", err)
		}
	}

	return nil
}

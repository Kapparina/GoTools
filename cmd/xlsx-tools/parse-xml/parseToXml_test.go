package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func createTestXlsx(fileName string, blankCols, blankRows bool) (filePath string, err error) {
	if suffix := filepath.Ext(fileName); suffix != ".xlsx" {
		fileName = strings.TrimSuffix(fileName, suffix) + ".xlsx"
	}
	var sheet int
	// Create a simple test xlsx file.
	excelFile := excelize.NewFile()

	// Create a test sheet.
	sheet, err = createTestSheet(excelFile, blankCols, blankRows)
	if err != nil {
		return
	}
	excelFile.SetActiveSheet(sheet)

	// Save the file.
	dir, _ := os.Getwd()
	filePath = filepath.Join(dir, fileName)
	if err = excelFile.SaveAs(filePath); err != nil {
		return
	}
	return
}

func createTestSheet(file *excelize.File, blankCols, blankRows bool) (sheet int, err error) {
	sheet, err = file.NewSheet("TestSheet")
	if err != nil {
		return
	}
	file.SetActiveSheet(sheet)
	for colIdx := 1; colIdx <= 10; colIdx++ {
		for rowIdx := 1; rowIdx <= 10; rowIdx++ {
			if rowIdx == 1 {
				cellName, _ := excelize.CoordinatesToCellName(colIdx, rowIdx)
				if err = file.SetCellValue("TestSheet", cellName, "Column"+cellName); err != nil {
					return
				}
			} else {
				cellName, _ := excelize.CoordinatesToCellName(colIdx, rowIdx)
				if err = file.SetCellValue("TestSheet", cellName, colIdx*rowIdx); err != nil {
					return
				}
			}
		}
	}
	if blankCols {
		for _, colIdx := range []int{2, 3, 4, 10} {
			for rowIdx := 2; rowIdx <= 10; rowIdx++ {
				cellName, _ := excelize.CoordinatesToCellName(colIdx, rowIdx)
				if err = file.SetCellValue("TestSheet", cellName, ""); err != nil {
					return
				}
			}
		}
	}
	if blankRows {
		for colIdx := 1; colIdx <= 10; colIdx++ {
			for rowIdx := 5; rowIdx <= 7; rowIdx++ {
				cellName, _ := excelize.CoordinatesToCellName(colIdx, rowIdx)
				if err = file.SetCellValue("TestSheet", cellName, ""); err != nil {
					return
				}
			}
		}
	}
	return
}

func createOutputXmlFile(fileName string, xmlData []byte) (string, error) {
	if suffix := filepath.Ext(fileName); suffix != ".xml" {
		fileName = strings.TrimSuffix(fileName, suffix) + ".xml"
	}
	dir, _ := os.Getwd()
	filePath := filepath.Join(dir, "test-results", fileName)

	// write output to xml file
	xmlFile, xmlFileErr := os.Create(filePath)
	if xmlFileErr != nil {
		return "", xmlFileErr
	}
	defer func() {
		if err := xmlFile.Close(); err != nil {
			xmlFileErr = err
		}
	}()
	_, writeErr := xmlFile.Write(xmlData)
	if writeErr != nil {
		return "", writeErr
	}
	return filePath, nil
}

func TestParseXlsxFile(t *testing.T) {
	// Define test files.
	type testFile struct {
		name      string
		blankCols bool
		blankRows bool
	}
	testFiles := []testFile{
		{"TestFile.xlsx", false, false},
		{"TestFileBlankRows", false, true},
		{"TestFileBlankCols.xls", true, false},
		{"TestFileBlankRowsAndCols.xls", true, true},
	}
	// Define test cases.
	tests := []struct {
		name        string
		filePath    string
		targetSheet string
		wantErr     bool
	}{
		{
			name:        "Valid File & Sheet",
			targetSheet: "TestSheet",
			wantErr:     false,
		},
		{
			name:        "Invalid File",
			filePath:    "InvalidPath",
			targetSheet: "TestSheet",
			wantErr:     true,
		},
		{
			name:        "Invalid Sheet",
			targetSheet: "InvalidSheet",
			wantErr:     true,
		},
	}
	// Now start with the testing.
	for fileNum, file := range testFiles {
		// Create test files - also cheekily testing the logic in createTestXlsx.
		filePath, err := createTestXlsx(file.name, file.blankCols, file.blankRows)
		if err != nil {
			t.Errorf("Error creating test file: %v", err)
		}
		t.Logf("Test file %d: %s", fileNum, filePath)
		for _, tt := range tests {
			if len(tt.filePath) < 1 {
				tt.filePath = filePath
			}
			// Clean up the test file when done.
			t.Run(tt.name, func(t *testing.T) {
				output, err := parseXlsxFile(tt.filePath, tt.targetSheet)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseXlsxFile() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !tt.wantErr {
					// Check if the output is not empty.
					if len(output) < 1 {
						t.Errorf("parseXlsxFile() output is empty")
					}
					outputPath, err := createOutputXmlFile(file.name, output)
					if err != nil {
						t.Errorf("Error creating output file: %v", err)
					}
					t.Logf("Output file: %s", outputPath)
				}
			})
		}
		if err = os.Remove(filePath); err != nil {
			t.Errorf("Error removing test file: %v", err)
		}
	}
}

// This program was designed to be used with the .NET Framework.
// It is designed to accommodate the parsing of a .xlsx file into a .NET DataTable object.
package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"flag"
	"os"
	"strings"

	. "GoTools/pkg/helpers"
	"github.com/xuri/excelize/v2"
)

type DataColumn struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type DataRow struct {
	Columns []DataColumn `xml:",any"`
}

type DataTable struct {
	Rows []DataRow `xml:"Row"`
}

// getInput retrieves user input for the file path and sheet name.
// It uses command line flags to get the user input, and falls back to standard input if no arguments are provided.
// The function trims any leading/trailing whitespace from the file path.
// It returns the file path, sheet name, and any input error encountered.
func getInput() (filePath, sheetName string, inputErr error) {
	flag.StringVar(&filePath, "path", "", "The path to the .xlsx file to parse")
	flag.StringVar(&sheetName, "sheet", "", "The name of the worksheet to parse")
	flag.Parse()

	if len(filePath) > 0 {
		filePath = strings.TrimSpace(filePath)
	} else {
		pipeInput, pipeErr := os.Stdin.Stat()
		if pipeErr != nil {
			inputErr = pipeErr
		}
		if pipeInput.Mode()&os.ModeNamedPipe != 0 {
			reader := bufio.NewReader(os.Stdin)
			input, bufferErr := reader.ReadString('\n')
			if bufferErr != nil {
				inputErr = bufferErr
			} else {
				filePath = strings.TrimSpace(input)
			}
		}
	}
	return
}

func main() {
	processingErr := ErrMsg{Code: Success}
	defer func() {
		processingErr.Exit()
	}()
	filePath, sheetName, inputErr := getInput()
	// Get user input
	if inputErr != nil {
		processingErr = ErrMsg{Err: inputErr, Code: ErrStdin}
	}
	// Validate user input
	if len(filePath) < 1 {
		processingErr = ErrMsg{Code: ErrNoInput}
	}
	// Validate file path
	exists, pathErr := PathExists(filePath)
	if pathErr != nil || !exists {
		processingErr = ErrMsg{Err: pathErr, Code: ErrNoFile}
	}
	// Validate file type
	if !isXlsxFile(filePath) {
		processingErr = ErrMsg{
			Err:  errors.New("invalid file type"),
			Code: ErrInvalidFileType,
		}
	}
	// Parse the file as XML
	output, parseErr := parseXlsxFile(filePath, sheetName)
	if parseErr != nil {
		processingErr = ErrMsg{Err: parseErr, Code: ErrParse}
	} else {
		// Write the output to stdout
		_, writeErr := os.Stdout.Write(output)
		if writeErr != nil {
			processingErr = ErrMsg{Err: writeErr, Code: ErrStdout}
		}
	}
	// // write output to xml file
	// xmlFile, xmlFileErr := os.Create("output.xml")
	// if xmlFileErr != nil {
	//     processingErr = ErrMsg{Err: xmlFileErr, Code: ErrWriteFile}
	// }
	// defer func() {
	//     if err := xmlFile.Close(); err != nil {
	//         processingErr = ErrMsg{Err: err, Code: ErrWriteFile}
	//     }
	// }()
	// _, writeErr := xmlFile.Write(output)
	// if writeErr != nil {
	//     processingErr = ErrMsg{Err: writeErr, Code: ErrWriteFile}
	// }
}

// CheckExtension checks if the given file path has the specified extension.
// It adds a dot to the beginning of the extension if it's missing.
// Returns true if the file extension matches the specified extension, and false otherwise.
func isXlsxFile(path string) bool {
	return CheckExtension(path, ".xlsx")
}

func parseXlsxFile(path, targetSheet string) (output []byte, parseErr error) {
	// Open the .xlsx file
	file, openFileErr := excelize.OpenFile(path)
	if openFileErr != nil {
		return nil, openFileErr
	}
	defer func(file *excelize.File) {
		err := file.Close()
		if err != nil {
			parseErr = err
		}
	}(file)

	// Get the target sheet, or the default if no target was provided
	var rows *excelize.Rows
	var rowsErr error

	if len(targetSheet) > 1 {
		rows, rowsErr = file.Rows(targetSheet)
	} else {
		rows, rowsErr = file.Rows(file.GetSheetName(0))
	}
	if rowsErr != nil {
		return nil, rowsErr
	}
	// Marshal the data into XML
	xmlOutput, marshalErr := xml.MarshalIndent(buildDataTable(rows), "", "  ")
	if marshalErr != nil {
		return nil, marshalErr
	} else {
		output = xmlOutput
	}
	return output, nil
}

// cleanHeader takes a pointer to a string `header` as input and modifies it.
// It calls the FixXMLTags function to clean the `header`, replacing any invalid XML characters.
// The modified `header` is then assigned back to the original pointer.
// Example usage:
//
//	header := "<Hello World!>"
//	cleanHeader(&header)
//	fmt.Println(header)
//	// Output: "Hello World"
func cleanHeader(header *string) {
	newHeader := *header
	newHeader = FixXMLTags(newHeader)
	*header = newHeader
}

// buildDataTable takes an excelize.Rows pointer as input and converts it into a DataTable struct.
// It iterates over each row in the rows and converts each row into a DataRow struct.
// If the rows pointer is nil, it returns an empty DataTable struct.
// For the first row, it renames any duplicate headers using the RenameDuplicates function.
// It then calls the cleanHeader function to clean each header.
// For subsequent rows, it converts each column into a DataColumn struct and appends it to the DataRow struct.
// The DataRow struct is then appended to the Rows field of the DataTable struct.
// The function returns the populated DataTable struct.
func buildDataTable(rows *excelize.Rows) DataTable {
	var dataTable DataTable
	var headerRow []string
	var rowIndex int
	if rows == nil {
		return dataTable
	}
	for rows.Next() {
		columns, colErr := rows.Columns()
		if colErr != nil {
			return DataTable{}
		}
		if rowIndex == 0 {
			headerRow = RenameDuplicates(columns, false)
			for headerIndex := range headerRow {
				cleanHeader(&headerRow[headerIndex])
			}
		} else {
			var dataRow DataRow
			for columnIndex := range columns {
				columnName := headerRow[columnIndex]
				columnValue := ConvertToISO8601(columns[columnIndex])
				column := DataColumn{XMLName: xml.Name{Local: columnName}, Value: columnValue}
				dataRow.Columns = append(dataRow.Columns, column)
			}
			dataTable.Rows = append(dataTable.Rows, dataRow)
		}
		rowIndex++
	}
	return dataTable
}

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

	"GoTools/src/helpers"
	"github.com/tealeg/xlsx"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Column struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type DataRow struct {
	Columns []Column `xml:",any"`
}

type DataTable struct {
	Rows []DataRow `xml:"Row"`
}

// getInput retrieves user input for the file path and sheet name.
// It uses command line flags to get the user input, and falls back to standard input if no arguments are provided.
// The function trims any leading/trailing whitespace from the file path.
// It returns the file path, sheet name, and any input error encountered.
func getInput() (filePath string, sheetName string, inputErr error) {
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
	processingErr := helpers.ErrMsg{Code: helpers.Success}
	defer func() {
		processingErr.Exit()
	}()
	filePath, sheetName, inputErr := getInput()
	// Get user input
	if inputErr != nil {
		processingErr = helpers.ErrMsg{Err: inputErr, Code: helpers.ErrStdin}
	}
	// Validate user input
	if len(filePath) < 1 {
		processingErr = helpers.ErrMsg{Code: helpers.ErrNoInput}
	}
	// Validate file path
	exists, pathErr := helpers.PathExists(filePath)
	if pathErr != nil || !exists {
		processingErr = helpers.ErrMsg{Err: pathErr, Code: helpers.ErrNoFile}
	}
	// Validate file type
	if !isXlsxFile(filePath) {
		processingErr = helpers.ErrMsg{
			Err:  errors.New("invalid file type"),
			Code: helpers.ErrInvalidFileType,
		}
	}
	// Parse the file as XML
	output, parseErr := parseXlsxFile(filePath, sheetName)
	if parseErr != nil {
		processingErr = helpers.ErrMsg{Err: parseErr, Code: helpers.ErrParse}
	} else {
		// Write the output to stdout
		_, writeErr := os.Stdout.Write(output)
		if writeErr != nil {
			processingErr = helpers.ErrMsg{Err: writeErr, Code: helpers.ErrStdout}
		}
	}
}

// CheckExtension checks if the given file path has the specified extension.
// It adds a dot to the beginning of the extension if it's missing.
// Returns true if the file extension matches the specified extension, and false otherwise.
func isXlsxFile(path string) bool {
	return helpers.CheckExtension(path, ".xlsx")
}

// parseXlsxFile uses the xlsx package to parse an .xlsx file and extract data from a specified sheet.
// It returns the parsed data as an XML-encoded byte slice and any encountered errors.
// The path parameter specifies the path to the .xlsx file to parse.
// The targetSheet parameter specifies the name of the target sheet to parse. If no target sheet is provided,
// the function will use the first sheet in the file.
// The function assumes that the schema (column headers) is located in the first row of the target sheet.
// It cleans the column headers by removing spaces and replacing invalid characters with valid XML element names.
// The function loads the data row by row, skipping the first row (schema) and populates a DataTable struct.
// Finally, it marshals the DataTable struct into XML format and returns the byte slice representation.
// If any errors occur during the parsing or marshaling process, an error will be returned.
// Example usage:
//
//		 filePath := "C:/temp/Config.xlsx"
//	  sheetName := "Sheet1"
//	  output, parseErr := parseXlsxFile(filePath, sheetName)
//	  if parseErr == nil {
//		   // Write the output to stdout
//		   _, writeErr := os.Stdout.Write(output)
//		   if writeErr != nil {
//			   fmt.Println("Error writing output:", writeErr)
//		   }
//	  } else {
//		   fmt.Println("Error parsing .xlsx file:", parseErr)
//	  }
func parseXlsxFile(path string, targetSheet string) ([]byte, error) {
	// Open the .xlsx file
	file, openFileErr := xlsx.OpenFile(path)
	if openFileErr != nil {
		return nil, openFileErr
	}
	var dataTable DataTable
	// Get the target sheet, or the default if no target was provided
	var sheet *xlsx.Sheet
	if len(targetSheet) > 1 {
		sheet = file.Sheet[targetSheet]
	} else {
		sheet = file.Sheets[0]
	}
	// Get the schema (assuming it is in the first row)
	var schema []string
	// Load schema, cleaning column headers
	for _, cell := range sheet.Rows[0].Cells {
		// String manipulation for making it a valid XML element name
		columnName := cell.String()
		if strings.ContainsAny(columnName, " ") {
			columnName = cases.Title(language.English).String(columnName)
			columnName = strings.ReplaceAll(columnName, " ", "")
		}
		columnName = strings.ReplaceAll(columnName, "?", "")
		schema = append(schema, columnName)
	}
	// Load data
	for rowIndex, row := range sheet.Rows {
		if rowIndex == 0 { // Skip first (schema) row
			continue
		}
		var dataRow DataRow
		for columnIndex, cell := range row.Cells {
			columnName := schema[columnIndex]
			column := Column{XMLName: xml.Name{Local: columnName}, Value: cell.String()}
			dataRow.Columns = append(dataRow.Columns, column)
		}
		dataTable.Rows = append(dataTable.Rows, dataRow)
	}
	output, marshalErr := xml.MarshalIndent(dataTable, "", "  ")
	if marshalErr != nil {
		return nil, marshalErr
	}
	return output, nil
}

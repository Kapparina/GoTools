package main

import (
    "bufio"
    "encoding/csv"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// Package name was omitted intentionally
// The following constants represent error codes for specific operations.
// Success indicates a successful operation.
// ErrReadFile indicates an error while reading a file.
// ErrWriteFile indicates an error while writing to a file.
// ErrStdin indicates an error while reading from standard input.
// ErrNoInput indicates that there is no input provided.
// ErrNoFile indicates that the file does not exist.
// ErrNotCSV indicates that the file is not in CSV format.
const (
    Success      = 0
    ErrReadFile  = 1
    ErrWriteFile = 2
    ErrStdin     = 3
    ErrNoInput   = 4
    ErrNoFile    = 5
    ErrNotCSV    = 6
)

// errMsg is a custom error type that represents an error and its corresponding code.
// err is the error that occurred.
// code is the code associated with the error.
// Example usage:
// 	var processingErr errMsg = errMsg{nil, Success}
// 	defer func() {
// 		processingErr.Exit()
// 	}()
//
// 	// rest of the code...
type errMsg struct {
    err  error
    code int
}

// Exit terminates the program with the provided exit code and prints an error message if there is an error.
// If e.err is not nil, it prints "An error occurred: <error message>" before exiting.
// It uses defer to ensure that os.Exit is always called, even if an error occurs.
// Example usage:
//   var processingErr errMsg = errMsg{nil, Success}
//   defer func() {
//     processingErr.Exit()
//   }()
//   ...
//   processingErr = processCSV(...)
func (e errMsg) Exit() {
    defer os.Exit(e.code)
    if e.err != nil {
        fmt.Printf("An error occured!\nError Code: %d\nDetail: %v\n", e.code, e.err)
    }
}

// main is the entry point of the program.
func main() {
    startTime := time.Now()
    defer fmt.Printf("Total time: %s\n", time.Since(startTime))

    var processingErr errMsg = errMsg{nil, Success}
    defer func() {
        processingErr.Exit()
    }()
    filePathPtr := flag.String("path", "", "CSV file path")
    flag.Parse()
    pipeInput, _ := os.Stdin.Stat()

    if pipeInput.Mode()&os.ModeNamedPipe != 0 {
        reader := bufio.NewReader(os.Stdin)
        input, inputErr := reader.ReadString('\n')
        if inputErr != nil {
            processingErr = errMsg{inputErr, ErrStdin}
        }
        processingErr = processCSV(strings.TrimSpace(input))
    } else if *filePathPtr != "" {
        processingErr = processCSV(*filePathPtr)
    } else {
        processingErr = errMsg{
            fmt.Errorf("no CSV path provided from pipe nor --path flag"),
            ErrNoInput,
        }
    }
}

// pathExists checks if a path exists or not.
func pathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    } else if os.IsNotExist(err) {
        return false, nil
    } else {
        return false, err
    }
}

// isCSV checks if the given file path has a .csv extension
func isCSV(path string) bool {
    return filepath.Ext(path) == ".csv"
}

// processCSV is a function that processes a CSV file located at the given file path.
// It performs the following steps:
// - Checks if the file exists. If not, it returns an error with code ErrNoFile.
// - Checks if the file is a CSV file. If not, it returns an error with code ErrNotCSV.
// - Reads the contents of the CSV file into a 2D string slice ([][]string) using the readCSV function.
// - Writes the CSV data back to the same file using the writeCSV function.
// - If any errors occur during the process, it returns an error with the appropriate code.
// - If the process is successful, it returns an empty error with code Success.
//
// Parameters:
// - filePath: The file path of the CSV file to process.
//
// Returns:
// - An instance of errMsg, which contains either an error and an error code, or a nil error and Success code.
func processCSV(filePath string) errMsg {
    if exists, _ := pathExists(filePath); !exists {
        return errMsg{fmt.Errorf("file '%s' does not exist", filePath), ErrNoFile}
    }
    if !isCSV(filePath) {
        return errMsg{fmt.Errorf("file '%s' is not a CSV file", filePath), ErrNotCSV}
    }
    csvData, csvReadErr := readCSV(filePath)

    if csvReadErr != nil {
        return errMsg{csvReadErr, ErrReadFile}
    }
    outputErr := writeCSV(filePath, csvData)
    if outputErr != nil {
        return errMsg{outputErr, ErrWriteFile}
    }
    return errMsg{nil, Success}
}

// readCSV reads a CSV file from the given filePath and returns a two-dimensional slice of strings representing the contents of the file.
// The first slice represents the headers of the CSV file and subsequent slices represent each row of data.
// If there is an error reading the file or processing the CSV, an error is returned along with nil data.
// Parameters:
// - filePath: A string representing the path to the CSV file.
// Return Values:
// - [][]string: A two-dimensional slice of strings representing the contents of the CSV file.
// - error: An error, if any, encountered while reading the file or processing the CSV.
// Example Usage:
// csvData, csvReadErr := readCSV(filePath)
// if csvReadErr != nil {
//     // Handle error
// }
// // Process csvData
func readCSV(filePath string) ([][]string, error) {
    file, ioErr := os.Open(filePath)
    if ioErr != nil {
        return nil, ioErr
    }
    defer func(file *os.File) {
        err := file.Close()
        if err != nil {
            log.Println(err)
        }
    }(file)
    reader := csv.NewReader(file)
    headers, headerErr := processHeaders(reader)

    if headerErr != nil {
        return nil, headerErr
    }
    records, recordErr := reader.ReadAll()
    if recordErr != nil {
        return nil, recordErr
    }
    return append([][]string{headers}, records...), nil
}

// processHeaders reads the headers from the csv reader and
// renames any duplicate headers by appending a number to them.
// It returns the processed headers or an error if reading fails.
func processHeaders(reader *csv.Reader) ([]string, error) {
    headers, err := reader.Read()
    if err != nil {
        return nil, err
    }
    return renameDuplicateHeaders(headers), nil
}

// renameDuplicateHeaders takes an input slice of strings and renames any duplicate headers
// by appending a count to them. It returns the modified input slice.
//
// Each header in the input slice is checked against a map called counts. The map stores
// the count of each header occurrence. If a header occurs more than once, its count is
// incremented and the header is renamed by appending "_<count>" to it.
//
// After the renaming is done, the counts map is iterated to print a message for each header
// that had duplicates.
//
// Example usage:
//   headers := []string{"Name", "Age", "Name", "City", "Age"}
//   modifiedHeaders := renameDuplicateHeaders(headers)
//
// Output:
//   Header 'Name' was present 2 times
//   Header 'Age' was present 2 times
//   Header 'Name_2' was present 1 times
//   Header 'City' was present 1 times
//
//   The modifiedHeaders slice will be:
//   []string{"Name", "Age", "Name_2", "City", "Age_2"}
func renameDuplicateHeaders(input []string) []string {
    counts := make(map[string]int)

    for i, header := range input {
        counts[header]++
        if counts[header] > 1 {
            input[i] = fmt.Sprintf("%s_%d", header, counts[header])
        }
    }
    for header, count := range counts {
        if count > 1 {
            fmt.Printf("Header '%s' was present %d times\n", header, count)
        }
    }
    return input
}

// writeCSV writes the given data to a CSV file specified by the filePath.
// It returns an error if an error occurs during the file write process.
func writeCSV(filePath string, data [][]string) error {
    file, ioErr := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
    if ioErr != nil {
        return ioErr
    }
    defer func(file *os.File) {
        err := file.Close()
        if err != nil {
            log.Println(err)
        }
    }(file)

    writer := csv.NewWriter(file)
    defer writer.Flush()

    writeErr := writer.WriteAll(data)
    if writeErr != nil {
        return writeErr
    }
    return nil
}

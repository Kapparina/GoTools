package main

import (
    "bufio"
    "encoding/csv"
    "flag"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    . "GoTools/pkg/helpers"
)

// main is the entry point of the program.
func main() {
    startTime := time.Now()
    defer fmt.Printf("Total time: %s\n", time.Since(startTime))

    processingErr := ErrMsg{Code: Success}
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
            processingErr = ErrMsg{Err: inputErr, Code: ErrStdin}
        }
        processingErr = processCSV(strings.TrimSpace(input))
    } else if *filePathPtr != "" {
        processingErr = processCSV(*filePathPtr)
    } else {
        processingErr = ErrMsg{
            Err:  fmt.Errorf("no CSV path provided from pipe nor --path flag"),
            Code: ErrNoInput,
        }
    }
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
// - An instance of helpers.ErrMsg, which contains either an error and an error code, or a nil error and Success code.
func processCSV(filePath string) ErrMsg {
    if exists, _ := PathExists(filePath); !exists {
        return ErrMsg{Err: fmt.Errorf("file '%s' does not exist", filePath), Code: ErrNoFile}
    }
    if !CheckExtension(filePath, ".csv") {
        return ErrMsg{
            Err:  fmt.Errorf("file '%s' is not a CSV file", filePath),
            Code: ErrInvalidFileType,
        }
    }
    csvData, csvReadErr := readCSV(filePath)

    if csvReadErr != nil {
        return ErrMsg{Err: csvReadErr, Code: ErrReadFile}
    }
    outputErr := writeCSV(filePath, csvData)
    if outputErr != nil {
        return ErrMsg{Err: outputErr, Code: ErrWriteFile}
    }
    return ErrMsg{Code: Success}
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
//
//	if csvReadErr != nil {
//	    // Handle error
//	}
//
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
    return RenameDuplicates(headers), nil
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

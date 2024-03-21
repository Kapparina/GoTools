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

func processCSV(path string) ErrMsg {
	if exists, _ := PathExists(path); !exists {
		return ErrMsg{Err: fmt.Errorf("file '%s' does not exist", path), Code: ErrNoFile}
	}
	if !CheckExtension(path, ".csv") {
		return ErrMsg{
			Err:  fmt.Errorf("file '%s' is not a CSV file", path),
			Code: ErrInvalidFileType,
		}
	}
	tempFile, ioErr := readWriteCsv(path)
	if ioErr != nil {
		return ErrMsg{Err: ioErr, Code: ErrReadWrite}
	}
	removeErr := os.Remove(path)
	if removeErr != nil {
		return ErrMsg{Err: removeErr, Code: ErrWriteFile}
	}
	log.Printf("Removed original file: '%s'\n", path)
	moveErr := MoveFile(tempFile, path)
	if moveErr != nil {
		return ErrMsg{Err: moveErr, Code: ErrMoveFile}
	}
	log.Printf("'%s' successfully replaced by '%s'", filepath.Base(path), filepath.Base(tempFile))
	return ErrMsg{Code: Success}
}

func readWriteCsv(path string) (string, error) {
	originalCsv, readErr := os.Open(path)
	if readErr != nil {
		return "", readErr
	}
	defer func(originalCsv *os.File) {
		err := originalCsv.Close()
		if err != nil {
			log.Println(err)
		}
	}(originalCsv)
	tempCsv, tempErr := os.CreateTemp("", fmt.Sprintf("*_%s", filepath.Base(path)))
	if tempErr != nil {
		return tempCsv.Name(), tempErr
	}
	defer func(tempCsv *os.File) {
		err := tempCsv.Close()
		if err != nil {
			log.Println(err)
		}
	}(tempCsv)
	log.Printf("Created temp file: '%s'\n", tempCsv.Name())

	reader := csv.NewReader(originalCsv)
	writer := csv.NewWriter(tempCsv)
	defer writer.Flush()

	log.Printf("Reading from: '%s'\n", originalCsv.Name())
	log.Printf("Writing to: '%s'\n", tempCsv.Name())
	lineCount := 0
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if lineCount == 0 {
			record = RenameDuplicates(record, true)
		}
		writeErr := writer.Write(record)
		if writeErr != nil {
			return tempCsv.Name(), writeErr
		}
		lineCount++
	}
	log.Printf("Successfully wrote amended headers plus original contents to: '%s'\n", tempCsv.Name())
	return tempCsv.Name(), nil
}

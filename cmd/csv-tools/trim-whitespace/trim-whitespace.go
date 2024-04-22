package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "GoTools/pkg/helpers"
	"github.com/charmbracelet/log"
)

func main() {
	log.SetLevel(log.DebugLevel)
	startTime := time.Now()

	processingErr := ErrMsg{Code: Success}
	defer func() {
		log.Debug(
			"DONE!",
			"time",
			time.Since(startTime),
		)
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
	log.Info(
		"Removed original file",
		"file", path,
	)
	moveErr := MoveFile(tempFile, path)
	if moveErr != nil {
		return ErrMsg{Err: moveErr, Code: ErrMoveFile}
	}
	log.Info(
		"Successfully amended file",
		"original", filepath.Base(path),
		"amended", filepath.Base(tempFile),
	)
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
			log.Error(err)
		}
	}(originalCsv)
	tempCsv, tempErr := os.CreateTemp("", fmt.Sprintf("*_%s", filepath.Base(path)))
	if tempErr != nil {
		return tempCsv.Name(), tempErr
	}
	defer func(tempCsv *os.File) {
		err := tempCsv.Close()
		if err != nil {
			log.Error(err)
		}
	}(tempCsv)
	log.Info("Created temp file", "file", tempCsv.Name())

	reader := csv.NewReader(originalCsv)
	writer := csv.NewWriter(tempCsv)
	defer writer.Flush()

	log.Info("ORIGINAL", "file", originalCsv.Name())
	log.Info("AMENDED", "file", tempCsv.Name())
	lineCount := 0
	for {
		record, err := reader.Read()
		newRecord := make([]string, len(record))
		if err != nil {
			break
		}
		for i, field := range record {
			newRecord[i] = strings.TrimSpace(field)
		}
		writeErr := writer.Write(newRecord)
		if writeErr != nil {
			return tempCsv.Name(), writeErr
		}
		lineCount++
	}
	log.Info("Trimmed whitespace successfully", "file", tempCsv.Name(), "lines", lineCount)
	return tempCsv.Name(), nil
}

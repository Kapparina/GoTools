package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

var (
	verbose bool
	dirPath string
)

type TargetFile struct {
	Path string
	Err  error
}

func handleError(target *TargetFile, err error, errChan chan<- *TargetFile) {
	target.Err = err
	errChan <- target
}

func getXmlEncoderDecoder(file *os.File) (*xml.Encoder, *xml.Decoder, *bytes.Buffer) {
	// Create a new buffered reader from the file
	reader := bufio.NewReader(file)

	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent(" ", "\t")

	decoder := xml.NewDecoder(reader)

	return encoder, decoder, &buf
}

func formatXmlFile(target *TargetFile, errChan chan<- *TargetFile, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(target.Path)
	if err != nil {
		handleError(target, err, errChan)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			handleError(target, err, errChan)
		}
	}()

	encoder, decoder, buf := getXmlEncoderDecoder(file)

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			handleError(target, err, errChan)
			return
		}
		if t == nil {
			break
		}
		if err := encoder.EncodeToken(t); err != nil {
			handleError(target, err, errChan)
			return
		}
	}

	if err := encoder.Flush(); err != nil {
		handleError(target, err, errChan)
		return
	}

	outFile, err := os.Create(target.Path)
	if err != nil {
		handleError(target, err, errChan)
		return
	}

	defer func(outFile *os.File) {
		if err := outFile.Close(); err != nil {
			handleError(target, err, errChan)
		}
	}(outFile)

	writer := bufio.NewWriter(outFile)
	if _, err = writer.Write(buf.Bytes()); err != nil {
		handleError(target, err, errChan)
		return
	}

	if err = writer.Flush(); err != nil {
		handleError(target, err, errChan)
		return
	}

	errChan <- target
}

func parseArgs() error {
	flag.StringVar(&dirPath, "path", "", "Path to directory containing XML files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
		log.SetCallerFormatter(log.LongCallerFormatter)
		log.SetReportCaller(true)
	}
	if len(dirPath) == 0 {
		log.Error("Enter either an absolute path to a directory or a specific XML file")
		flag.Usage()
		return errors.New("no path provided")
	}
	return nil
}

func main() {
	defer func(startTime time.Time) {
		log.Debug("TIME!", "execution time", time.Since(startTime))
	}(time.Now())

	if argErr := parseArgs(); argErr != nil {
		log.Error(argErr)
		return
	}
	xmlFiles, dirErr := prepareXMLFiles()
	if dirErr != nil {
		log.Error(dirErr)
		return
	}

	processFilesConcurrently(xmlFiles)
}

func prepareXMLFiles() ([]TargetFile, error) {
	var xmlFiles []TargetFile
	dirInfo, dirErr := os.Stat(dirPath)
	if dirErr != nil {
		return nil, dirErr
	}
	dirPath = filepath.Clean(dirPath)

	if dirInfo.IsDir() {
		log.Info("Processing XML files in directory", "path", dirPath)
		files, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".xml") {
				xmlFiles = append(
					xmlFiles,
					TargetFile{Path: filepath.Join(dirPath, file.Name())},
				)
			}
		}
	} else if strings.HasSuffix(dirPath, ".xml") {
		log.Info("Processing XML file", "path", dirPath)
		xmlFiles = append(xmlFiles, TargetFile{Path: dirPath})
	}
	return xmlFiles, nil
}

func processFilesConcurrently(xmlFiles []TargetFile) {
	result := make(chan *TargetFile, len(xmlFiles))
	defer func(res chan *TargetFile) {
		for r := range res {
			if r.Err != nil {
				log.Error(
					"Error formatting XML file",
					"file name", filepath.Base(r.Path),
					"error", r.Err,
				)
			} else {
				log.Info(
					"XML file formatted successfully",
					"file name", filepath.Base(r.Path),
				)
			}
		}
	}(result)

	var wg sync.WaitGroup
	wg.Add(len(xmlFiles))

	for i := 0; i < len(xmlFiles); i++ {
		log.Info(
			"Processing file",
			"file name", filepath.Base(xmlFiles[i].Path),
		)
		go formatXmlFile(&xmlFiles[i], result, &wg)
	}
	go func() {
		wg.Wait()
		close(result)
	}()
}

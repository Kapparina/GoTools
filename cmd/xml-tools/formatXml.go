package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
)

type TargetFile struct {
	Path string
	Err  error
}

func formatXmlFile(target *TargetFile, errChan chan<- *TargetFile, wg *sync.WaitGroup) {
	defer wg.Done()

	reader, openErr := os.Open(target.Path)
	if openErr != nil {
		target.Err = openErr
		errChan <- target
		return
	}
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			target.Err = err
			errChan <- target
			return
		}
	}(reader)

	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent(" ", "\t")

	decoder := xml.NewDecoder(reader)
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				target.Err = err
				errChan <- target
				return
			}
		}
		if t == nil {
			break
		}
		err = encoder.EncodeToken(t)
		if err != nil {
			target.Err = err
			errChan <- target
			return
		}
	}
	if err := encoder.Flush(); err != nil {
		target.Err = err
		errChan <- target
		return
	}

	if err := os.WriteFile(target.Path, buf.Bytes(), 0644); err != nil {
		target.Err = err
		errChan <- target
		return
	}
	errChan <- target
	return
}

func main() {
	var dirPath string
	var xmlFiles []TargetFile

	flag.StringVar(&dirPath, "path", "", "Path to directory containing XML files")
	flag.Parse()

	if len(dirPath) == 0 {
		log.Error("Enter either an absolute path to a directory or a specific XML file")
		flag.Usage()
		os.Exit(1)
	}

	dirInfo, dirErr := os.Stat(dirPath)
	if dirErr != nil {
		log.Fatal(dirErr)
	}

	dirPath = filepath.Clean(dirPath)

	if dirInfo.IsDir() {
		log.Info("Processing XML files in directory", "path", dirPath)
		files, err := os.ReadDir(dirPath)
		if err != nil {
			log.Fatal(err)
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

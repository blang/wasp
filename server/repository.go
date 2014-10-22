package server

import (
	"bufio"
	"code.google.com/p/lzma"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/blang/wasp/filescan"
)

const SCAN_WORKER_COUNT = 10
const BUILD_WORKER_COUNT = 10

type Repository struct {
	srcPath      string
	destPath     string
	SrcDirectory *filescan.Directory
	// DestDirectory *filescan.Directory
}

func NewRepository(srcPath, destPath string) *Repository {
	return &Repository{
		srcPath:  srcPath,
		destPath: destPath,
	}
}

func (r *Repository) Scan() error {
	var err error
	r.SrcDirectory, err = filescan.Scan(r.srcPath, SCAN_WORKER_COUNT)
	if err != nil {
		return err
	}
	// r.DestDirectory, err = filescan.Scan(r.destPath, SCAN_WORKER_COUNT)
	// if err != nil {
	// 	return err
	// }
	return nil
}

type prefixFile struct {
	prefix string
	file   *filescan.File
}

type MultiError struct {
	Errors []error
}

func (m MultiError) Error() string {
	var errors []string
	for _, err := range m.Errors {
		errors = append(errors, err.Error())
	}
	return fmt.Sprintf("%d Errors occurred: %s", len(m.Errors), strings.Join(errors, ", "))
}

func (m *MultiError) Add(err error) {
	m.Errors = append(m.Errors, err)
}

func (r *Repository) Build() error {
	fileCh := make(chan prefixFile, 100)
	multiError := &MultiError{}
	wg := sync.WaitGroup{}
	wg.Add(BUILD_WORKER_COUNT)
	for i := 0; i < BUILD_WORKER_COUNT; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case preFile, ok := <-fileCh:
					if !ok {
						return
					}
					inputPath := filepath.Join(preFile.prefix, preFile.file.Name)
					outputPath := filepath.Join(r.destPath, preFile.file.Hash+".xz")
					err := compress(inputPath, outputPath)
					if err != nil {
						multiError.Add(fmt.Errorf("Could not compress file %s to %s", inputPath, outputPath))
					}
				}
			}
		}()
	}

	filescan.Walk(filepath.Dir(r.srcPath), r.SrcDirectory, filescan.FileWalkFunc(func(prefix string, file *filescan.File) {
		fileCh <- prefixFile{
			prefix: prefix,
			file:   file,
		}
	}), nil)
	close(fileCh)
	wg.Wait()
	if len(multiError.Errors) > 0 {
		return multiError
	}
	return nil
}

func compress(inputFile, outputFile string) error {
	outFile, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer outFile.Close()

	bufW := bufio.NewWriter(outFile)

	w := lzma.NewWriter(bufW)
	fileR, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer fileR.Close()
	bufR := bufio.NewReader(fileR)
	io.Copy(w, bufR)
	w.Close()
	err = bufW.Flush()
	if err != nil {
		return err
	}
	return nil
}

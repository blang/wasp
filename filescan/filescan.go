package filescan

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type fileInfo struct {
	path string
	file *File
}

func Scan(root string, workerCount int) (*Directory, error) {
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("Not a directory: %s", root)
	}
	db := NewDirectory(filepath.Base(root))
	dirDB := make(map[string]*Directory)
	// dirDB[filepath.Dir(root)] = db
	dirDB[root] = db

	fileCh := make(chan fileInfo)
	closeCh := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {

		go func() {
			b := make([]byte, 4069*1024)
			defer wg.Done()
			for {
				select {
				case fileInfo, ok := <-fileCh:
					if !ok {
						return
					}
					f, err := os.Open(fileInfo.path)
					if err != nil {
						log.Printf("File error on %s: %s", fileInfo.path, err)
						continue
					}
					hash := md5.New()
					bufr := bufio.NewReader(f)
					for {
						read, err := bufr.Read(b)
						if err != nil || read <= 0 {
							break
						}
						hash.Write(b[:read])
					}

					fileInfo.file.Hash = fmt.Sprintf("%x", hash.Sum(nil))
				case <-closeCh:
					return
				}
			}
		}()
	}

	walkFn := filepath.WalkFunc(func(path string, info os.FileInfo, err error) error {
		baseName := filepath.Base(path)
		parentDir := filepath.Dir(path)
		if info == nil {
			return fmt.Errorf("No fileInfo: %s", path)
		}
		if info.IsDir() {
			dir := NewDirectory(baseName)
			parent, ok := dirDB[parentDir]
			if !ok {
				return nil
				// return fmt.Errorf("Directory has no registered parent: %s", path)
			}
			parent.Directories[baseName] = dir
			dirDB[path] = dir
			return nil
		}

		file := NewFile(baseName, info.Size(), info.ModTime().Unix())
		parent, ok := dirDB[parentDir]
		if !ok {
			return fmt.Errorf("File has no registered parent: %s, parentDir: %s, dirDB: %s", path, parentDir, dirDB)
		}
		parent.Files[baseName] = file

		fileCh <- fileInfo{
			path: path,
			file: file,
		}
		return nil
	})
	err = filepath.Walk(root, walkFn)
	if err != nil {
		close(closeCh)
		return nil, err
	}
	close(fileCh)
	wg.Wait()
	return db, nil
}

type FileWalkFunc func(prefix string, file *File)
type DirWalkFunc func(prefix string, dir *Directory)

func Walk(prefix string, root *Directory, fileWalkFn FileWalkFunc, dirWalkFn DirWalkFunc) {
	curPrefix := filepath.Join(prefix, root.Name)
	if fileWalkFn != nil {
		for _, file := range root.Files {
			fileWalkFn(curPrefix, file)
		}
	}

	for _, dir := range root.Directories {
		if dirWalkFn != nil {
			dirWalkFn(curPrefix, dir)
		}
		Walk(curPrefix, dir, fileWalkFn, dirWalkFn)
	}
}

package filescan

import (
	"fmt"
	"path/filepath"
	"testing"
)

const FIXTURE_DIR = "/tmp/wasp"

func createTestStructure(rootDir string) *Directory {
	root := NewDirectory(filepath.Base(rootDir))
	dirA := NewDirectory("a")
	dirAA := NewDirectory("a")
	dirAB := NewDirectory("b")
	dirA.AddDirectory(dirAA)
	dirA.AddDirectory(dirAB)
	root.AddDirectory(dirA)

	fileEmpty := NewFile("empty.txt", 0, 0)
	fileEmpty.Hash = "d41d8cd98f00b204e9800998ecf8427e"
	dirA.AddFile(fileEmpty)

	file1 := NewFile("1.txt", 6, 0)
	file1.Hash = "3e7705498e8be60520841409ebc69bc1"
	dirAA.AddFile(file1)

	file2 := NewFile("2.txt", 6, 0)
	file2.Hash = "126a8a51b9d1bbd07fddc65819a542c3"
	dirAB.AddFile(file2)
	return root
}

func TestScan(t *testing.T) {
	rootDir, err := Scan(FIXTURE_DIR, 4)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	testDir := createTestStructure(FIXTURE_DIR)
	if err := compareDirectory(testDir, rootDir); err != nil {
		t.Fatalf("Error: %s", err)
	}
}

func compareDirectory(testDir, realDir *Directory) error {
	if testDir.Name != realDir.Name {
		return fmt.Errorf("Dirname, expected %s, got %s", testDir.Name, realDir.Name)
	}
	if len(testDir.Directories) != len(realDir.Directories) {
		return fmt.Errorf("Subdirectories, expected %s, got %s", testDir.Directories, realDir.Directories)
	}
	if len(testDir.Files) != len(realDir.Files) {
		return fmt.Errorf("Files, expected %s, got %s", testDir.Files, realDir.Files)
	}
	for testPath, testFile := range testDir.Files {
		realFile, ok := realDir.Files[testPath]
		if !ok {
			return fmt.Errorf("File %s not found", testPath)
		}
		if testFile.Name != realFile.Name {
			return fmt.Errorf("Filename, expected %s, got %s", testFile.Name, realFile.Name)
		}
		if testFile.Size != realFile.Size {
			return fmt.Errorf("Filesize %s, expected %d, got %d", testFile.Name, testFile.Size, realFile.Size)
		}
		if testFile.Hash != realFile.Hash {
			return fmt.Errorf("Hash %s, expected %s, got %s", testFile.Name, testFile.Hash, realFile.Hash)
		}
	}
	for subTestDirName, subTestDir := range testDir.Directories {
		subRealDir, ok := realDir.Directories[subTestDirName]
		if !ok {
			return fmt.Errorf("Subdirectory %s not found", subTestDirName)
		}
		if err := compareDirectory(subTestDir, subRealDir); err != nil {
			return err
		}
	}
	return nil
}

func TestWalk(t *testing.T) {
	testDir := createTestStructure(FIXTURE_DIR)

	Walk("", testDir, FileWalkFunc(func(prefix string, file *File) {
		t.Logf("File: %s", filepath.Join(prefix, file.Name))
	}), DirWalkFunc(func(prefix string, dir *Directory) {
		t.Logf("Dir: %s", filepath.Join(prefix, dir.Name))
	}))
}

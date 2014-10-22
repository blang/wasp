package filescan

type File struct {
	Name         string
	Size         int64
	LastModified int64
	Hash         string
}

func NewFile(name string, size int64, lastmodified int64) *File {
	return &File{
		Name:         name,
		Size:         size,
		LastModified: lastmodified,
	}
}

type Directory struct {
	Name        string
	Directories map[string]*Directory
	Files       map[string]*File
}

func NewDirectory(name string) *Directory {
	return &Directory{
		Name:        name,
		Directories: make(map[string]*Directory),
		Files:       make(map[string]*File),
	}
}

func (dir *Directory) AddDirectory(subDir *Directory) {
	dir.Directories[subDir.Name] = subDir
}

func (dir *Directory) AddFile(file *File) {
	dir.Files[file.Name] = file
}

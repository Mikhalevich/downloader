package downloader

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileStorer struct {
	FolderName string
	FileName   string
}

func NewFileStorer(folder string) *FileStorer {
	return &FileStorer{
		FolderName: folder,
		FileName:   "",
	}
}

func (fs *FileStorer) Store(bytes []byte) error {
	if fs.FileName == "" {
		return errors.New("Invalid file name")
	}

	if fs.FolderName != "" {
		if err := os.MkdirAll(fs.FolderName, os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create(filepath.Join(fs.FolderName, fs.FileName))
	if err != nil {
		return err
	}
	defer file.Close()

	if len(bytes) > 0 {
		file.Write(bytes)
	}

	return nil
}

func (fs *FileStorer) Get() ([]byte, error) {
	file, err := os.Open(filepath.Join(fs.FolderName, fs.FileName))
	if err != nil {
		return []byte(""), err
	}

	return ioutil.ReadAll(file)
}

func (fs *FileStorer) GetFileName() string {
	return fs.FileName
}

func (fs *FileStorer) SetFileName(fileName string) {
	fs.FileName = fileName
}

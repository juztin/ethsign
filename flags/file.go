package flags

import (
	"errors"
	"os"
)

type FileFlag struct {
	path string
}

func (f *FileFlag) String() string {
	return f.path
}

func (f *FileFlag) Set(value string) error {
	if value == "" {
		f.path = "."
		return nil
	}
	info, err := os.Stat(value)
	if err != nil {
		return errors.Unwrap(err)
	} else if info.IsDir() {
		return errors.New("file is a directory")
	} else {
		f.path = value
	}
	return err
}

type FilePathFlag struct {
	FileFlag
}

func (f *FilePathFlag) Set(value string) error {
	if value == "" {
		f.path = "."
		return nil
	}
	_, err := os.Stat(value)
	if err != nil {
		return errors.Unwrap(err)
	}
	f.path = value
	return err
}

func File(value string) FileFlag {
	return FileFlag{value}
}

func FilePath(value string) FilePathFlag {
	f := FileFlag{value}
	return FilePathFlag{f}
}

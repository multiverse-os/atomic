package atomicio

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrAlreadyCommitted = errors.New("[error] file already committed")

type File struct {
	*os.File
	originalName string
	closeFunc    func(*File) error
	isClosed     bool // if true, temporary file has been closed, but not renamed
	isCommitted  bool // if true, the file has been successfully committed
}

func makeTempName(originalName, prefix string) (tempName string, err error) {
	originalName = filepath.Clean(originalName)
	if len(originalName) == 0 || originalName[len(originalName)-1] == filepath.Separator {
		return "", os.ErrInvalid
	}
	// Generate 10 random bytes.
	// This gives 80 bits of entropy, good enough
	// for making temporary file name unpredictable.
	var randomBytes [10]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}
	name := prefix + "-" + strings.ToLower(base32.StdEncoding.EncodeToString(randomBytes[:])) + ".tmp"
	return filepath.Join(filepath.Dir(originalName), name), nil
}

// Create creates a temporary file in the same directory as filename,
// which will be renamed to the given filename when calling Commit.
// TODO Instead of having an infinite loop should always opt for a time
// out option to make software more resiliant to edge conditions
func Create(name string, permissions os.FileMode) (*File, error) {
	for {
		if tempName, err := makeTempName(name, "temp"); err != nil {
			return nil, err
		} else {
			if file, err := os.OpenFile(tempName, os.O_RDWR|os.O_CREATE|os.O_EXCL, permissions); err != nil {
				if os.IsExist(err) {
					continue
				}
				return nil, err
			} else {
				return &File{
					File:         file,
					originalName: name,
					closeFunc:    closeUncommitted,
				}, nil
			}
		}
	}
}

func (self *File) Name() string {
	if self.isCommitted {
		return self.OriginalName()
	} else {
		return self.File.Name()
	}
}

func (self *File) OriginalName() string {
	return self.originalName
}

func (self *File) Close() error {
	return self.closeFunc(self)
}

func closeUncommitted(file *File) error {
	if err := file.File.Close(); err != nil {
		os.Remove(file.Name())
		return err
	}
	if err := os.Remove(file.Name()); err != nil {
		return err
	}
	return nil
}

func closeAfterFailedRename(file *File) error {
	file.closeFunc = closeAgainError
	return os.Remove(file.Name())
}

func closeAgainError(file *File) error {
	return os.ErrInvalid
}

func (self *File) Commit() error {
	if self.isCommitted {
		return ErrAlreadyCommitted
	}
	if !self.isClosed {
		if err := self.Sync(); err != nil {
			return err
		}
		if err := self.File.Close(); err != nil {
			return err
		}
		self.isClosed = true
	}
	if err := rename(self.Name(), self.originalName); err != nil {
		self.closeFunc = closeAfterFailedRename
		return err
	}
	self.isCommitted = true
	fmt.Println("self.Name:", self.Name())
	return nil
}

func WriteFile(name string, data []byte, permissions os.FileMode) error {
	if file, err := Create(name, permissions); err != nil {
		return err
	} else {
		defer file.Close()
		if bytesWritten, err := file.Write(data); err != nil {
			return err
		} else {
			if err == nil && bytesWritten < len(data) {
				return io.ErrShortWrite
			}
			return file.Commit()
		}
	}
}

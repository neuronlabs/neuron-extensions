package osfs

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/filestore"
)

var _ filestore.File = &File{}

// File is an abstraction that defines per-filesystem file.
type File struct {
	name, directory, bucket, version string
	path                             string
	// modifiedAt is the last modification date.
	modifiedAt  time.Time
	permissions os.FileMode
	size        int64
	closed      bool
	f           *os.File
	w           bytes.Buffer
}

// Name implements files.File.
func (f *File) Name() string {
	return f.name
}

// Bucket implements files.File interface.
func (f *File) Bucket() string {
	return f.bucket
}

// Directory implements files.File.
func (f *File) Directory() string {
	return f.directory
}

// Version implements files.File.
func (f *File) Version() string {
	return f.version
}

// ModifiedAt implements files.File.
func (f *File) ModifiedAt() time.Time {
	return f.modifiedAt
}

// Size implements files.File interface.
func (f *File) Size() int64 {
	return f.size
}

// Open implements files.File.
func (f *File) Open(_ context.Context) error {
	if f.f != nil {
		return filestore.ErrAlreadyOpened
	}
	f.closed = false

	// Open
	file, err := os.Open(f.path)
	if err == nil {
		f.f = file
		info, err := file.Stat()
		if err != nil {
			return errors.Wrap(filestore.ErrFileStore, err.Error())
		}
		if info.IsDir() {
			return filestore.ErrFileIsDir
		}
		f.modifiedAt = info.ModTime()
		f.size = info.Size()
		f.permissions = info.Mode().Perm()
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return filestore.ErrNotExists
	}
	if errors.Is(err, os.ErrPermission) {
		return filestore.ErrPermission
	}
	return errors.Wrap(filestore.ErrInternal, err.Error())
}

// Close implements files.File interface.
func (f *File) Close(_ context.Context) error {
	if f.f == nil {
		if f.closed {
			return filestore.ErrClosed
		}
		return filestore.ErrNotOpened
	}
	file := f.f
	f.f = nil
	f.closed = true
	err := file.Close()
	if err != nil {
		if errors.Is(err, os.ErrClosed) {
			return filestore.ErrClosed
		}
		return filestore.ErrInternal
	}
	return nil
}

// Read implements files.File and os.Reader interface.
func (f *File) Read(data []byte) (int, error) {
	if f.f == nil {
		if f.closed {
			return 0, filestore.ErrClosed
		}
		return 0, filestore.ErrNotOpened
	}
	n, err := f.f.Read(data)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return n, err
		}
		if errors.Is(err, os.ErrClosed) {
			return n, filestore.ErrClosed
		}
		return n, errors.Wrap(filestore.ErrInternal, err.Error())
	}
	return n, nil
}

// Write implements files.File and io.Writer interface.
func (f *File) Write(data []byte) (int, error) {
	n, err := f.w.Write(data)
	if err == nil {
		return 0, nil
	}
	if errors.Is(err, io.EOF) {
		return n, err
	}
	if errors.Is(err, io.ErrShortWrite) {
		return n, err
	}
	return n, errors.Wrap(filestore.ErrInternal, err.Error())
}

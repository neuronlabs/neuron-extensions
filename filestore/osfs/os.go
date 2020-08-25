package osfs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"

	"github.com/neuronlabs/neuron/errors"
	"github.com/neuronlabs/neuron/filestore"
)

var _ filestore.Store = &OsStore{}

// OsStore implements files.Store interface.
type OsStore struct {
	Options *Options
}

// New creates new FileSystem with given options.
func New(options ...Option) (*OsStore, error) {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	if o.RootDirectory != "" {
		info, err := os.Lstat(o.RootDirectory)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			if errors.Is(err, os.ErrPermission) {
				return nil, errors.Wrap(filestore.ErrPermission, err.Error())
			}
			return nil, errors.Wrap(filestore.ErrInternal, err.Error())
		}
		if info != nil && !info.IsDir() {
			return nil, errors.Wrap(filestore.ErrStore, "provided root directory is a file")
		}
	}
	if o.RootDirectory == "" {
		o.RootDirectory = "."
	}
	var err error
	o.RootDirectory, err = filepath.Abs(o.RootDirectory)
	if err != nil {
		return nil, errors.Wrapf(filestore.ErrStore, "root directory abs failed: %v", err.Error())
	}
	if o.DefaultDirectory != "" {
		dir := filepath.Join(o.RootDirectory, o.DefaultDirectory)
		info, err := os.Lstat(dir)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			if errors.Is(err, os.ErrPermission) {
				return nil, errors.Wrap(filestore.ErrPermission, err.Error())
			}
			return nil, errors.Wrap(filestore.ErrInternal, err.Error())
		}
		if info != nil && !info.IsDir() {
			return nil, errors.Wrap(filestore.ErrStore, "provided default directory is a file")
		}
	}
	if o.DirectoryPermissions == 0 {
		o.DirectoryPermissions = 0755
	}
	return &OsStore{Options: o}, nil
}

// NewFile creates new file skeleton for given name and options.
func (f *OsStore) NewFile(_ context.Context, name string, options ...filestore.FileOption) (filestore.File, error) {
	dir, fName := filepath.Split(name)
	if dir != "" || fName == "" {
		return nil, errors.Wrap(filestore.ErrFileName, "provided path is not simple file name")
	}
	o := &filestore.FileOptions{}
	for _, option := range options {
		option(o)
	}
	if o.Bucket == "" && f.Options.DefaultBucket != "" {
		o.Bucket = f.Options.DefaultBucket
	}
	if o.Directory != "" {
		if strings.Contains(o.Directory, ".versions") {
			return nil, errors.Wrap(filestore.ErrFileName, "filename directory contains prohibited directory .versions")
		}
	}
	if o.Directory == "" && f.Options.DefaultDirectory != "" {
		o.Directory = f.Options.DefaultDirectory
	}
	file := &File{
		name:      fName,
		directory: o.Directory,
		bucket:    o.Bucket,
	}
	return file, nil
}

// PutFile sets up the file in the store.
func (f *OsStore) PutFile(ctx context.Context, file filestore.File, options ...filestore.PutOption) error {
	osFile, ok := file.(*File)
	if !ok {
		return errors.Wrap(filestore.ErrFileStore, "provided file type is not osfs.File")
	}
	o := &filestore.PutOptions{}
	for _, option := range options {
		option(o)
	}
	if f.Options.FileVersions {
		return f.putFileV(ctx, osFile, o)
	}
	return f.putFileNV(ctx, osFile, o)
}

func (f *OsStore) putFileV(ctx context.Context, file *File, _ *filestore.PutOptions) error {
	dir := f.fileVersionsDirPath(file)
	versionsLock := filepath.Join(dir, "/.versions.lock")
	// Set up write/update lock on the lock file.
	lock := flock.New(versionsLock)
	for {
		ok, err := lock.TryLockContext(ctx, time.Microsecond*20)
		if err != nil {
			// If the lock does not exists create full path for given lock.
			if errors.Is(err, os.ErrNotExist) {
				if err = os.MkdirAll(dir, f.Options.DirectoryPermissions); err != nil {
					if errors.Is(err, os.ErrPermission) {
						return errors.Wrap(filestore.ErrPermission, err.Error())
					}
					if !errors.Is(err, os.ErrExist) {
						return errors.Wrap(filestore.ErrFileStore, err.Error())
					}
					err = nil
				}
			} else {
				if errors.Is(err, os.ErrPermission) {
					return errors.Wrap(filestore.ErrPermission, err.Error())
				}
				return errors.Wrap(filestore.ErrFileStore, err.Error())
			}
		}
		if ok {
			break
		}
	}
	defer lock.Unlock()

	if err := f.setNextFileVersion(file, dir); err != nil {
		return err
	}
	path := f.fileVersionPath(file)
	of, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return errors.Wrap(filestore.ErrPermission, err.Error())
		}
		return errors.Wrap(filestore.ErrFileStore, err.Error())
	}
	if _, err = file.w.WriteTo(of); err != nil {
		of.Close()
		return err
	}
	of.Close()
	// Clear previous symbolic link if exists and set it up to
	if err = f.clearPreviousFileVersion(file); err != nil {
		return err
	}
	// Create new symbolic link to the non versioned path that points to the newest version path.
	if err = os.Symlink(path, f.fileNonVersionPath(file)); err != nil {
		if errors.Is(err, os.ErrExist) {
			return errors.Wrap(filestore.ErrExists, err.Error())
		}
		if errors.Is(err, os.ErrPermission) {
			return errors.Wrap(filestore.ErrPermission, err.Error())
		}
		return errors.Wrap(filestore.ErrFileStore, err.Error())
	}
	return nil
}

func (f *OsStore) putFileNV(_ context.Context, file *File, o *filestore.PutOptions) error {
	path := f.fileNonVersionPath(file)
	flags := os.O_WRONLY | os.O_CREATE
	if o.Overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}
	of, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return errors.Wrap(filestore.ErrPermission, err.Error())
		}
		if errors.Is(err, os.ErrExist) {
			return errors.Wrap(filestore.ErrExists, err.Error())
		}
		return errors.Wrap(filestore.ErrFileStore, err.Error())
	}
	defer of.Close()
	if _, err = file.w.WriteTo(of); err != nil {
		return err
	}
	return nil
}

// GetFile implements files.Store.
func (f *OsStore) GetFile(_ context.Context, name string, options ...filestore.GetOption) (filestore.File, error) {
	dir, fName := filepath.Split(name)
	if dir != "" || fName == "" {
		return nil, errors.Wrap(filestore.ErrFileName, "provided path is not simple file name")
	}
	o := &filestore.GetOptions{}
	for _, option := range options {
		option(o)
	}
	if o.Bucket == "" && f.Options.DefaultBucket != "" {
		o.Bucket = f.Options.DefaultBucket
	}
	if !f.Options.FileVersions && o.Version != "" {
		return nil, filestore.ErrVersionsNotAllowed
	}
	if o.Directory != "" {
		if strings.Contains(o.Directory, ".versions") {
			return nil, errors.Wrap(filestore.ErrFileName, "filename directory contains prohibited directory .versions")
		}
	}
	if o.Directory == "" && f.Options.DefaultDirectory != "" {
		o.Directory = f.Options.DefaultDirectory
	}
	file := &File{
		name:      fName,
		version:   o.Version,
		bucket:    o.Bucket,
		directory: o.Directory,
	}
	if f.Options.FileVersions {
		if file.version == "" {
			latest, err := f.getFileLatestVersion(file)
			if err != nil {
				return nil, err
			}
			file.version = latest
		}
		file.path = f.fileVersionPath(file)
	} else {
		file.path = f.fileNonVersionPath(file)
	}
	info, err := os.Lstat(file.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, filestore.ErrNotExists
		}
		if errors.Is(err, os.ErrPermission) {
			return nil, errors.Wrap(filestore.ErrPermission, err.Error())
		}
		return nil, errors.Wrap(filestore.ErrInternal, err.Error())
	}
	if info != nil && info.IsDir() {
		return nil, errors.Wrap(filestore.ErrFileName, "provided name is a directory")
	}
	if info != nil {
		file.modifiedAt = info.ModTime()
		file.size = info.Size()
		file.permissions = info.Mode().Perm()
	}
	return file, nil
}

// ListFiles lists all files in provided directory - filtered with optional options. Implements files.Store.
func (f *OsStore) ListFiles(_ context.Context, dir string, options ...filestore.ListOption) ([]filestore.File, error) {
	o := &filestore.ListOptions{}
	for _, option := range options {
		option(o)
	}
	if o.Bucket == "" && f.Options.DefaultBucket != "" {
		o.Bucket = f.Options.DefaultBucket
	}

	if dir == "" && f.Options.DefaultDirectory != "" {
		dir = f.Options.DefaultDirectory
	}
	if strings.Contains(dir, ".versions") {
		return nil, errors.Wrap(filestore.ErrFileName, "filename directory contains prohibited directory .versions")
	}

	fullPath := filepath.Join(f.Options.RootDirectory, o.Bucket, dir)

	var fs []filestore.File
	if o.Limit == 0 {
		o.Limit = -1
	}

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if path == fullPath {
			return nil
		}
		if info.IsDir() {
			return filepath.SkipDir
		}

		if o.Offset != 0 {
			o.Offset--
			return nil
		}
		if o.Limit == 0 {
			return nil
		}
		o.Limit--
		if o.Extension != "" && o.Extension != filepath.Ext(path) {
			return nil
		}
		_, name := filepath.Split(path)
		fs = append(fs, &File{
			name:        name,
			path:        filepath.Join(f.Options.RootDirectory, path),
			size:        info.Size(),
			permissions: info.Mode().Perm(),
			modifiedAt:  info.ModTime(),
		})
		return nil
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, filestore.ErrNotExists
		}
		if errors.Is(err, os.ErrPermission) {
			return nil, filestore.ErrPermission
		}
		return nil, errors.Wrap(filestore.ErrInternal, err.Error())
	}
	return fs, nil
}

// DeleteFile implements files.Store.
func (f *OsStore) DeleteFile(ctx context.Context, file filestore.File) error {
	if f.Options.FileVersions {
		return f.deleteFileV(ctx, file)
	}
	return f.deleteFileNV(file)
}

func (f *OsStore) deleteFileV(ctx context.Context, file filestore.File) error {
	dir := f.fileVersionsDirPath(file)
	versionsLock := filepath.Join(dir, "/.versions.lock")
	// Set up write/update lock on the lock file.
	lock := flock.New(versionsLock)
	for {
		ok, err := lock.TryLockContext(ctx, time.Microsecond*20)
		if err != nil {
			// If the lock does not exists create full path for given lock.
			if errors.Is(err, os.ErrNotExist) {
				return filestore.ErrNotExists
			}
			if errors.Is(err, os.ErrPermission) {
				return errors.Wrap(filestore.ErrPermission, err.Error())
			}
			return errors.Wrap(filestore.ErrFileStore, err.Error())
		}
		if ok {
			break
		}
	}
	defer lock.Unlock()

	if file.Version() == "" {
		// Delete all file versions.
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return filepath.SkipDir
			}
			if strings.HasSuffix(path, "/.versions.lock") {
				return nil
			}
			if err = os.Remove(path); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return filestore.ErrNotExists
				}
				if errors.Is(err, os.ErrPermission) {
					return filestore.ErrPermission
				}
				return errors.Wrap(filestore.ErrFileStore, err.Error())
			}
			return nil
		})
		if err != nil {
			return err
		}
		if err := os.Remove(f.fileNonVersionPath(file)); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return filestore.ErrNotExists
			}
			if errors.Is(err, os.ErrPermission) {
				return filestore.ErrPermission
			}
			return errors.Wrap(filestore.ErrFileStore, err.Error())
		}
		return nil
	}
	// Remove from versions list.
	versionsFileName := filepath.Join(dir, "/.versions")
	versionsFile, err := os.OpenFile(versionsFileName, os.O_RDWR, 0664)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return filestore.ErrNotExists
		}
		if errors.Is(err, os.ErrPermission) {
			return filestore.ErrPermission
		}
		return errors.Wrap(filestore.ErrFileStore, err.Error())
	}
	buf := &bytes.Buffer{}
	thisLine := -1
	var (
		lineCount  int
		prev, line string
	)
	fileName := file.Version() + filepath.Ext(file.Name())
	for {
		_, err = fmt.Fscanln(versionsFile, &line)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			versionsFile.Close()
			return errors.Wrap(filestore.ErrInternal, err.Error())
		}
		lineCount++
		if line == fileName {
			// Don't write this line - we want to clear it from file.
			thisLine = lineCount
		} else {
			_, err = fmt.Fprintln(buf, line)
			if err != nil {
				versionsFile.Close()
				return errors.Wrap(filestore.ErrInternal, err.Error())
			}
			prev = line
		}
	}
	versionsFile.Close()
	if thisLine == -1 {
		// This file version doesn't exists.
		return errors.Wrap(filestore.ErrNotExists, "file version doesn't exists")
	}
	versionsFile, err = os.OpenFile(versionsFileName, os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return filestore.ErrNotExists
		}
		if errors.Is(err, os.ErrPermission) {
			return filestore.ErrPermission
		}
		return errors.Wrap(filestore.ErrFileStore, err.Error())
	}
	_, err = buf.WriteTo(versionsFile)
	if err != nil {
		versionsFile.Close()
		return err
	}
	versionsFile.Close()

	// If this version is the latest and there is still some previous version - set it as the main version now, and recreate symbolic link.
	if thisLine == lineCount {
		// If there were no previous version - clear symbolic link.
		err = os.Remove(f.fileNonVersionPath(file))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			if errors.Is(err, os.ErrPermission) {
				return errors.Wrap(filestore.ErrPermission, err.Error())
			}
			return errors.Wrap(filestore.ErrInternal, err.Error())
		}
		if prev != "" {
			// Set the symbolic link to the previous file version.
			prevPath := filepath.Join(f.fileVersionsDirPath(file), prev)
			if err = os.Symlink(prevPath, f.fileNonVersionPath(file)); err != nil {
				if errors.Is(err, os.ErrPermission) {
					return errors.Wrap(filestore.ErrPermission, err.Error())
				}
				return errors.Wrap(filestore.ErrInternal, err.Error())
			}
		}
	}
	// Remove this file
	if err = os.Remove(f.fileVersionPath(file)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Wrap(filestore.ErrNotExists, err.Error())
		}
		if errors.Is(err, os.ErrPermission) {
			return errors.Wrap(filestore.ErrPermission, err.Error())
		}
		return errors.Wrap(filestore.ErrInternal, err.Error())
	}

	return nil
}

func (f *OsStore) deleteFileNV(file filestore.File) error {
	path := f.fileNonVersionPath(file)
	err := os.Remove(path)
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return filestore.ErrNotExists
	}
	if errors.Is(err, os.ErrPermission) {
		return filestore.ErrPermission
	}
	return errors.Wrap(filestore.ErrFileStore, err.Error())
}

// Type implements files.Store.
func (f *OsStore) Type() string {
	return "os"
}

func (f *OsStore) fileNonVersionPath(file filestore.File) string {
	return filepath.Join(f.fileDirPath(file), file.Name())
}

func (f *OsStore) fileDirPath(file filestore.File) string {
	return filepath.Join(f.Options.RootDirectory, file.Bucket(), file.Directory())
}

// os.RootDirectory + file.Bucket + file.Directory + file_Name
func (f *OsStore) fileVersionsDirPath(file filestore.File) string {
	return filepath.Join(f.Options.RootDirectory, file.Bucket(), file.Directory(), ".files-versions", file.Name())
}

func (f *OsStore) fileVersionPath(file filestore.File) string {
	return filepath.Join(f.fileVersionsDirPath(file), file.Version()+filepath.Ext(file.Name()))
}

func (f *OsStore) getFileLatestVersion(file *File) (string, error) {
	dir := f.fileVersionsDirPath(file)
	versionsLock := filepath.Join(dir, "/.versions.lock")
	lock := flock.New(versionsLock)
	for {
		ok, err := lock.TryRLock()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", filestore.ErrNotExists
			}
			if errors.Is(err, os.ErrPermission) {
				return "", errors.Wrap(filestore.ErrPermission, err.Error())
			}
			return "", errors.Wrap(filestore.ErrFileStore, err.Error())
		}
		if ok {
			break
		}
	}
	defer lock.Unlock()
	versionsFileName := filepath.Join(dir, "/.versions")
	versionsFile, err := os.OpenFile(versionsFileName, os.O_RDONLY, 0664)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", filestore.ErrNotExists
		}
		if errors.Is(err, os.ErrPermission) {
			return "", filestore.ErrPermission
		}
		return "", errors.Wrapf(filestore.ErrFileStore, "%v", err)
	}
	defer versionsFile.Close()
	var last, line string
	for {
		last = line
		_, err = fmt.Fscan(versionsFile, &line)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", errors.Wrap(filestore.ErrFileStore, err.Error())
		}
	}
	last = strings.TrimSuffix(last, filepath.Ext(file.name))
	return last, nil
}

func (f *OsStore) setNextFileVersion(file *File, dir string) error {
	file.version = f.nextVersion(file.name)
	versionsFileName := filepath.Join(dir, "/.versions")
	versionsFile, err := os.OpenFile(versionsFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return filestore.ErrPermission
		}
		return errors.Wrapf(filestore.ErrFileStore, "%v", err)
	}
	defer versionsFile.Close()

	_, err = fmt.Fprintln(versionsFile, file.version+filepath.Ext(file.name))
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return errors.Wrap(filestore.ErrPermission, err.Error())
		}
		return errors.Wrap(filestore.ErrFileStore, err.Error())
	}
	return nil
}

func (f *OsStore) clearPreviousFileVersion(file *File) error {
	link := f.fileNonVersionPath(file)
	err := os.Remove(link)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		if errors.Is(err, os.ErrPermission) {
			return errors.Wrap(filestore.ErrPermission, err.Error())
		}
		return errors.Wrap(filestore.ErrFileStore, err.Error())
	}
	return nil
}

func (f *OsStore) nextVersion(name string) string {
	h := sha256.New()
	h.Write([]byte(name))
	tb, _ := time.Now().MarshalBinary()
	h.Write(tb)
	return hex.EncodeToString(h.Sum(nil))
}

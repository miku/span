// Package atomicfile writes files atomically. Content goes to a temporary
// file in the target directory, which replaces the target on Commit.
package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
)

// File is a temporary file that replaces the file at its target path when
// committed.
type File struct {
	*os.File
	path string
	done bool
}

// Create creates a temporary file next to path. Call Commit to move it into
// place. Close without Commit removes the temporary file, so a deferred Close
// cleans up on error.
func Create(path string, perm os.FileMode) (*File, error) {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+"-tmp-*")
	if err != nil {
		return nil, err
	}
	if err := f.Chmod(perm); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	return &File{File: f, path: path}, nil
}

// Commit flushes the file to disk, closes it and renames it to the target
// path.
func (f *File) Commit() error {
	if f.done {
		return errors.New("atomicfile: already committed or closed")
	}
	f.done = true
	err := errors.Join(f.File.Sync(), f.File.Close())
	if err == nil {
		err = os.Rename(f.Name(), f.path)
	}
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}

// Close discards the temporary file, unless it has been committed.
func (f *File) Close() error {
	if f.done {
		return nil
	}
	f.done = true
	f.File.Close()
	return os.Remove(f.Name())
}

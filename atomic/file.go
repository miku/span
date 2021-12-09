package atomic

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// Compress and return path to compressed file.
func Compress(filename string) (string, error) {
	tf, err := ioutil.TempFile("", "span-atomic-")
	if err != nil {
		return "", err
	}
	defer tf.Close()
	zw := gzip.NewWriter(tf)
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(zw, f); err != nil {
		return "", err
	}
	if err := zw.Flush(); err != nil {
		return "", err
	}
	return tf.Name(), nil
}

// WriteFile writes the data to a temp file and atomically move if everything else succeeds.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	dir, name := path.Split(filename)
	f, err := ioutil.TempFile(dir, name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if permErr := os.Chmod(f.Name(), perm); err == nil {
		err = permErr
	}
	if err == nil {
		err = os.Rename(f.Name(), filename)
	}
	// Any err should result in full cleanup.
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}

// Move moves a file from a source path to a destination path
// This must be used across the codebase for compatibility with Docker volumes
// and Golang (fixes Invalid cross-device link when using os.Rename)
func Move(src, dst string) error {
	sabs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	dabs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	if sabs == dabs {
		return nil
	}
	inf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inf.Close()
	dstDir := filepath.Dir(dst)
	if !exists(dstDir) {
		err = os.MkdirAll(dstDir, 0775)
		if err != nil {
			return err
		}
	}
	of, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer of.Close()
	_, err = io.Copy(of, inf)
	if err != nil {
		return err
	}
	return os.Remove(src)
}

// exists returns whether or not a file or path exists
func exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

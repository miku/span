package atomic

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	gzip "github.com/klauspost/pgzip"
)

// Compress and return path to compressed file.
func Compress(filename string) (string, error) {
	tf, err := ioutil.TempFile("", "span-atomic-")
	if err != nil {
		return "", err
	}
	zw := gzip.NewWriter(tf)
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(zw, f); err != nil {
		return "", err
	}
	if err := zw.Flush(); err != nil {
		return "", err
	}
	if err := tf.Close(); err != nil {
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

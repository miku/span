package atomic

import (
	"io/ioutil"
	"os"
	"path"
)

// WriteFileAtomic writes the data to a temp file and atomically move if everything else succeeds.
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

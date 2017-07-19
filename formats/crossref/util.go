package crossref

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// HashReader hashes the content of a reader. If reader is a seeker, we actually
// rewind to the beginning. The must not be manipulated elsewhere during this operation.
func HashReader(r io.Reader) (digest string, err error) {
	if s, ok := r.(io.Seeker); ok {
		if _, err := s.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// Harvest represents an API download.
type Harvest struct {
	r io.Reader
}

// Sum returns a SHA256 fingerprint of the content.
func (f *Harvest) Sum() (string, error) {
	return HashReader(f.r)
}

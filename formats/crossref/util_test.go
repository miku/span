package crossref

import (
	"strings"
	"testing"
)

func TestHashReader(t *testing.T) {
	digest, err := HashReader(strings.NewReader("abc"))
	if err != nil {
		t.Error(err)
	}
	x := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if digest != x {
		t.Errorf("got %s, want %s", digest, x)
	}
}

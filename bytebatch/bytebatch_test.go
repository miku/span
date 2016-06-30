package bytebatch

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
)

var errFake = errors.New("fake err")

func TestBatchError(t *testing.T) {
	r := strings.NewReader("hello\nworld\n")
	w := ioutil.Discard

	callCount := 0
	f := func(b []byte) ([]byte, error) {
		callCount++
		return b, errFake
	}
	p := NewLineProcessor(r, w, f)
	err := p.Run()
	if err != errFake {
		t.Errorf("p.Run() got %v, want %v", err, errFake)
	}
	if callCount != 2 {
		t.Errorf("callCount: got %v, want %v", callCount, 2)
	}
}

func TestBatch(t *testing.T) {
	r := strings.NewReader("hello\nworld\n")
	w := ioutil.Discard

	callCount := 0
	f := func(b []byte) ([]byte, error) {
		callCount++
		return b, nil
	}
	p := NewLineProcessor(r, w, f)
	err := p.Run()
	if err != nil {
		t.Errorf("p.Run() got %v, want %v", err, nil)
	}
	if callCount != 2 {
		t.Errorf("callCount: got %v, want %v", callCount, 2)
	}
}

package atomic

import (
	"bytes"
	"os"
	"testing"
)

func test(t *testing.T, dir, prefix string) {
	t.Parallel()

	tmpfile, err := os.CreateTemp(dir, prefix)
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()

	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}

	defer os.Remove(name)
	f, err := New(name, os.FileMode(0666))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("foo")); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		t.Fatal("did not expect file to exist")
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(name); err != nil {
		t.Fatalf("expected file to exist: %s", err)
	}
}

func TestCurrentDir(t *testing.T) {
	cwd, _ := os.Getwd()
	test(t, cwd, "atomicfile-current-dir-")
}

func TestRootTmpDir(t *testing.T) {
	test(t, "/tmp", "atomicfile-root-tmp-dir-")
}

func TestDefaultTmpDir(t *testing.T) {
	test(t, "", "atomicfile-default-tmp-dir-")
}

func TestAbort(t *testing.T) {
	contents := []byte("the answer is 42")
	t.Parallel()
	tmpfile, err := os.CreateTemp("", "atomicfile-abort-")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()
	if _, err := tmpfile.Write(contents); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	f, err := New(name, os.FileMode(0666))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("foo")); err != nil {
		t.Fatalf("%v", err)
	}
	if err := f.Abort(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(name); err != nil {
		t.Fatalf("expected file to exist: %s", err)
	}
	actual, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(contents, actual) {
		t.Fatalf(`did not find expected "%s" instead found "%s"`, contents, actual)
	}
}

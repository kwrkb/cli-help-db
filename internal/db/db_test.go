package db

import (
	"testing"
)

func TestWriteAndRead(t *testing.T) {
	d := New(t.TempDir())

	if err := d.Write("docker", "Usage: docker [OPTIONS] COMMAND"); err != nil {
		t.Fatal(err)
	}

	text, err := d.Read("docker")
	if err != nil {
		t.Fatal(err)
	}
	if text != "Usage: docker [OPTIONS] COMMAND" {
		t.Errorf("got %q", text)
	}
}

func TestRead_NotExist(t *testing.T) {
	d := New(t.TempDir())

	text, err := d.Read("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if text != "" {
		t.Errorf("expected empty string, got %q", text)
	}
}

func TestHas(t *testing.T) {
	d := New(t.TempDir())

	if d.Has("foo") {
		t.Error("expected Has=false before write")
	}

	if err := d.Write("foo", "help text here\nmore text"); err != nil {
		t.Fatal(err)
	}

	if !d.Has("foo") {
		t.Error("expected Has=true after write")
	}
}

func TestList(t *testing.T) {
	d := New(t.TempDir())

	for _, name := range []string{"curl", "aws", "kubectl"} {
		if err := d.Write(name, "help for "+name); err != nil {
			t.Fatal(err)
		}
	}

	names, err := d.List()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"aws", "curl", "kubectl"}
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, names[i], want[i])
		}
	}
}

func TestList_EmptyDir(t *testing.T) {
	d := New(t.TempDir())
	names, err := d.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty list, got %v", names)
	}
}

func TestList_NonexistentDir(t *testing.T) {
	d := New("/nonexistent/db/path")
	names, err := d.List()
	if err != nil {
		t.Fatal(err)
	}
	if names != nil {
		t.Errorf("expected nil, got %v", names)
	}
}

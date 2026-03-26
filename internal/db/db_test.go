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

func TestNameToKey(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"curl", "curl"},
		{"docker container ls", "docker__container__ls"},
		{"kubectl get pods", "kubectl__get__pods"},
	}
	for _, tt := range tests {
		if got := NameToKey(tt.name); got != tt.want {
			t.Errorf("NameToKey(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestKeyToName(t *testing.T) {
	tests := []struct {
		key, want string
	}{
		{"curl", "curl"},
		{"docker__container__ls", "docker container ls"},
		{"kubectl__get__pods", "kubectl get pods"},
	}
	for _, tt := range tests {
		if got := KeyToName(tt.key); got != tt.want {
			t.Errorf("KeyToName(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestSubcommand_WriteReadHas(t *testing.T) {
	d := New(t.TempDir())

	name := "docker container ls"
	text := "Usage: docker container ls [OPTIONS]"

	if err := d.Write(name, text); err != nil {
		t.Fatal(err)
	}

	if !d.Has(name) {
		t.Error("expected Has=true for subcommand")
	}

	got, err := d.Read(name)
	if err != nil {
		t.Fatal(err)
	}
	if got != text {
		t.Errorf("got %q, want %q", got, text)
	}
}

func TestSubcommand_List(t *testing.T) {
	d := New(t.TempDir())

	commands := []string{"curl", "docker", "docker container ls", "docker image ls"}
	for _, name := range commands {
		if err := d.Write(name, "help for "+name); err != nil {
			t.Fatal(err)
		}
	}

	names, err := d.List()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"curl", "docker", "docker container ls", "docker image ls"}
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, names[i], want[i])
		}
	}
}

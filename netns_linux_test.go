package netns

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestNewNamedExisting(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, err := Get()
	if err != nil {
		t.Fatal(err)
	}
	defer origns.Close()
	defer func() {
		current, err := Get()
		if err != nil {
			t.Fatal(err)
		}
		defer current.Close()
		if !current.Equal(origns) {
			if err := Set(origns); err != nil {
				t.Fatalf("failed to restore namespace: %v", err)
			}
			t.Error("failed creation changed the calling thread's namespace")
		}
	}()

	dir := t.TempDir()
	namedPath := filepath.Join(dir, "existing")
	if err := os.WriteFile(namedPath, []byte("existing namespace"), 0o600); err != nil {
		t.Fatal(err)
	}
	ns, err := newNamed(namedPath)
	if ns.IsOpen() {
		ns.Close()
		t.Error("failed creation returned an open namespace handle")
	}
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("newNamed error = %v; want file exists", err)
	}
	got, err := os.ReadFile(namedPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "existing namespace" {
		t.Fatalf("existing namespace file changed: %q", got)
	}
}

func TestNewNamedExistingFallback(t *testing.T) {
	dir := t.TempDir()
	primaryDir := filepath.Join(dir, "primary")
	fallbackDir := filepath.Join(dir, "fallback")
	for _, d := range []string{primaryDir, fallbackDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	namedPath := filepath.Join(fallbackDir, "existing")
	if err := os.WriteFile(namedPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	// Both directories exist, but only the fallback directory contains the name.
	resolvedPath := getNetnsPathForNameInDirs("existing", primaryDir, fallbackDir)
	if resolvedPath != namedPath {
		t.Fatalf("resolved path = %q; want %q", resolvedPath, namedPath)
	}
	ns, err := newNamed(resolvedPath)
	if ns.IsOpen() {
		ns.Close()
		t.Error("failed creation returned an open namespace handle")
	}
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("newNamed error = %v; want file exists", err)
	}
	if _, err := os.Stat(filepath.Join(primaryDir, "existing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("duplicate primary namespace file: %v", err)
	}
}

func TestGetNewSetDelete(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, err := Get()
	if err != nil {
		t.Fatal(err)
	}
	newns, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if origns.Equal(newns) {
		t.Fatal("New ns failed")
	}
	if err := Set(origns); err != nil {
		t.Fatal(err)
	}
	if err := newns.Close(); err != nil {
		t.Error("Failed to close ns", err)
	}
	if newns.IsOpen() {
		t.Fatal("newns still open after close", newns)
	}
	ns, err := Get()
	if err != nil {
		t.Fatal(err)
	}
	if !ns.Equal(origns) {
		t.Fatal("Reset ns failed", origns, newns, ns)
	}
}

func TestNone(t *testing.T) {
	ns := None()
	if ns.IsOpen() {
		t.Fatal("None ns is open", ns)
	}
}

func TestThreaded(t *testing.T) {
	ncpu := runtime.GOMAXPROCS(-1)
	if ncpu < 2 {
		t.Skip("-cpu=2 or larger required")
	}

	// Lock this thread simply to ensure other threads get used.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	wg := &sync.WaitGroup{}
	for i := 0; i < ncpu; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			TestGetNewSetDelete(t)
		}()
	}
	wg.Wait()
}

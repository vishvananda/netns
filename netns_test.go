package netns

import (
	"runtime"
	"testing"
)

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
	newns.Close()
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

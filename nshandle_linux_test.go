package netns_test

import (
	"testing"

	"github.com/vishvananda/netns"
)

func TestNsHandle_Name(t *testing.T) {
	const expectedName = "some-ns-no-one-will-ever-use"

	// Create a new network namespace
	ns, err := netns.NewNamed(expectedName)
	if err != nil {
		t.Fatalf("Failed to create network namespace: %v", err)
	}

	t.Cleanup(func() {
		if err := ns.Close(); err != nil {
			t.Fatalf("Failed to close network namespace: %v", err)
		}
		if err := netns.DeleteNamed(expectedName); err != nil {
			t.Fatalf("Failed to delete network namespace: %v", err)
		}
	})

	// Get the name of the network namespace
	name, err := ns.Name()
	if err != nil {
		t.Fatalf("Failed to get network namespace name: %v", err)
	}
	if name != expectedName {
		t.Fatalf("Expected name %q, got %q", expectedName, name)
	}
}

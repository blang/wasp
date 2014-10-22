package server

import (
	"testing"
)

func TestRepo(t *testing.T) {
	r := NewRepository("/tmp/wasp", "/tmp/waspout")
	err := r.Scan()
	if err != nil {
		t.Fatalf("Error while scanning: %s", err)
	}

	err = r.Build()
	if err != nil {
		t.Fatalf("Error while building: %s", err)
	}
}

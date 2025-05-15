package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup code here (if needed)

	// Run tests
	code := m.Run()

	// Teardown code here (if needed)

	os.Exit(code)
}

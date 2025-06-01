package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCmdVersion(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Error: %+v", err)
	}
	fmt.Printf("Output: %s", buf.String())
}

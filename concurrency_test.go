package main

import (
	"slices"
	"testing"
)

func TestDoWork(t *testing.T) {
	var got []string
	doWork(&got)

	// useless since next test will also check the len
	gotLen := len(got)
	if gotLen != 4 {
		t.Errorf("log len %d, expected 4", gotLen)
	}

	want := []string{"start", "end", "cleanup B", "cleanup A"}
	if !slices.Equal(got, want) {
		t.Errorf("log slices not equal, got %q, expected %q", got, want)
	}
}

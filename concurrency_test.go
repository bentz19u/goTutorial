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
		t.Errorf("log slices not equal, got %v, expected %v", got, want)
	}
}

func TestSquareAll(t *testing.T) {
	tests := map[string]struct {
		input, want []int
	}{
		"Three numbers": {[]int{2, 6, 15}, []int{4, 36, 225}},
		"Empty":         {[]int{}, []int{}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := squareAll(test.input)

			if !slices.Equal(got, test.want) {
				t.Errorf("square numbers dont match, got %v, want %v", got, test.want)
			}
		})
	}
}

func TestCounter(t *testing.T) {
	got := counter()
	want := 10000

	if got != want {
		t.Errorf("counter don't match, got %d, want %d", got, want)
	}
}

func TestChannelCounter(t *testing.T) {
	got := channelCounter()
	want := 10000

	if got != want {
		t.Errorf("counter don't match, got %d, want %d", got, want)
	}
}

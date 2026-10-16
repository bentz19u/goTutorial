package main

import (
	"errors"
	"slices"
	"testing"
	"time"
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

func TestBufferedChannelCounter(t *testing.T) {
	got := bufferedChannelCounter()
	want := 10000

	if got != want {
		t.Errorf("counter don't match, got %d, want %d", got, want)
	}
}

func TestReceiveWithTimeout(t *testing.T) {
	tests := map[string]struct {
		value, want, wait, timeout int
		wantErr                    error
	}{"Return value": {10, 10, 1, 20, nil}, "Timeout": {10, 0, 20, 1, ErrTimeout}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// it's buffered so in case of no receiver (timeout case), it doesnt wait
			ch := make(chan int, 1)

			go func() {
				time.Sleep(time.Duration(test.wait) * time.Millisecond)
				ch <- test.value
			}()

			got, err := receiveWithTimeout(ch, time.Duration(test.timeout)*time.Millisecond)

			if !errors.Is(err, test.wantErr) {
				t.Errorf("failed error: got %v, want %v", err, test.wantErr)
			}

			if test.want != got {
				t.Errorf("failed value: got %d, want %d", got, test.want)
			}
		})
	}
}

// some tick can be missing, if the ticker trigger while still running, it's just ditched
// so we test against wantMin and wantMax to include lost ticks
func TestPollCounter(t *testing.T) {
	tests := map[string]struct {
		interval, duration, wantMin, wantMax int
	}{"Return 100": {1, 100, 90, 100}, "Return 50": {2, 100, 45, 50}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				time.Sleep(time.Duration(test.duration) * time.Millisecond)
				close(done)
			}()

			got := pollCounter(time.Duration(test.interval)*time.Millisecond, done)

			if test.wantMin > got || got > test.wantMax {
				t.Errorf("failed value: got %d, wantMin %d, wantMax %d", got, test.wantMin, test.wantMax)
			}
		})
	}
}

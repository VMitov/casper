package caspertest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func GetStdout(t *testing.T, f func()) string {
	old := os.Stdout // keep backup of the real stdout
	defer func() { os.Stdout = old }()
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(fmt.Errorf("An error wasn't expected: %v", err))
	}
	os.Stdout = w

	// execute the main function
	f()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	out := <-outC

	return out
}

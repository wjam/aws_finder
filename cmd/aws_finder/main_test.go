package main

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	os.Args = append([]string{os.Args[0]}, os.Args[3:]...)

	main()
}

func TestBashCompletion(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "completion", "bash") //go run . completion bash
	cmd.Env = append(cmd.Env, "GO_WANT_HELPER_PROCESS=1")

	contentChan := make(chan []byte, 1)
	go func() {
		content, err := cmd.CombinedOutput()
		require.NoError(t, err)

		contentChan <- content
	}()

	select {
	case content := <-contentChan:
		assert.NotEqual(t, "PASS", string(content))
	case <-time.After(3 * time.Second):
		assert.Fail(t, "Timed out waiting for main() to finish")
	}
}

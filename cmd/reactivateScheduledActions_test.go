package cmd

import (
	"testing"
)

/**
  This is a smoke test that only checks that the program does not panic.
*/
func Test_reEnableWorkflows(t *testing.T) {
	client := getGithubClient()
	reEnableWorkflows("testing-release-robot", client)
}

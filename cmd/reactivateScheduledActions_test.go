package cmd

import (
	"testing"
)

func Test_reEnableWorkflows(t *testing.T) {
	client := getGithubClient()
	reEnableWorkflows("cloudwatch-adapter", client)
}

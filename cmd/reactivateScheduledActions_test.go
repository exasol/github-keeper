package cmd

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ReEnableWorkflowsSuite struct {
	suite.Suite
}

func TestReEnableWorkflowsSuite(t *testing.T) {
	suite.Run(t, new(ReEnableWorkflowsSuite))
}

/**
  This is a smoke test that only checks that the program does not panic.
*/
func (suite *ReEnableWorkflowsSuite) Test_reEnableWorkflows() {
	reactivateScheduledActionsCmd.Run(reactivateScheduledActionsCmd, []string{"testing-release-robot"})
}

func (suite *ReEnableWorkflowsSuite) TestUnknownRepo() {
	client := getGithubClient()
	suite.PanicsWithValue("Failed to list the workflows of um-unknown-repo. Cause: GET https://api.github.com/repos/exasol/um-unknown-repo/actions/workflows?per_page=1000: 404 Not Found []", func() {
		reEnableWorkflows("um-unknown-repo", client)
	})
}

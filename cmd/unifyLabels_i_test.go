package cmd

import (
	"context"
	"testing"

	"github.com/google/go-github/v39/github"
	"github.com/stretchr/testify/suite"
)

type UnifyLabelsSuite struct {
	IntegrationTestSuite
}

func TestUnifyLabelsSuite(t *testing.T) {
	suite.Run(t, new(UnifyLabelsSuite))
}

func (suite *UnifyLabelsSuite) TestCreateLables() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	suite.runUnifyLabelCommand(true)
	label, _, err := suite.githubClient.Issues.GetLabel(context.Background(), suite.testOrg, suite.testRepo, "feature")
	suite.NoError(err)
	suite.Equal(*label.Name, "feature")
	suite.Equal(*label.Color, "88ee66")
}

func (suite *UnifyLabelsSuite) runUnifyLabelCommand(fix bool) {
	UnifyLabels(suite.testRepo, getGithubClient(), fix)
}

func (suite *UnifyLabelsSuite) TestRenameLabel() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	labelName := "blocked"
	issueName := "TestIssue"
	issue, _, err := suite.githubClient.Issues.Create(context.Background(), suite.testOrg, suite.testRepo, &github.IssueRequest{Title: &issueName, Labels: &[]string{labelName}})
	suite.NoError(err)
	suite.runUnifyLabelCommand(true) // should update label to blocked:yes
	updatedIssue, _, err := suite.githubClient.Issues.Get(context.Background(), suite.testOrg, suite.testRepo, *issue.Number)
	suite.NoError(err)
	var labelNames []string
	for _, label := range updatedIssue.Labels {
		labelNames = append(labelNames, *label.Name)
	}
	suite.Assert().Contains(labelNames, "blocked:yes")
}

func (suite *UnifyLabelsSuite) TestMigrateLabel() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	blockedLabel := "blocked"
	blockedYesLabel := "blocked:yes"
	_, _, err := githubClient.Issues.CreateLabel(context.Background(), suite.testOrg, suite.testRepo, &github.Label{Name: &blockedYesLabel})
	suite.NoError(err)
	issueName := "TestIssue"
	issue, _, err := githubClient.Issues.Create(context.Background(), suite.testOrg, suite.testRepo, &github.IssueRequest{Title: &issueName, Labels: &[]string{blockedLabel}})
	suite.NoError(err)
	suite.runUnifyLabelCommand(true)
	updatedIssue, _, err := githubClient.Issues.Get(context.Background(), suite.testOrg, suite.testRepo, *issue.Number)
	suite.NoError(err)
	var labelNames []string
	for _, label := range updatedIssue.Labels {
		labelNames = append(labelNames, *label.Name)
	}
	suite.Assert().Contains(labelNames, blockedYesLabel)
	suite.Assert().NotContains(labelNames, blockedLabel)
}

func (suite *UnifyLabelsSuite) TestChangeColor() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	featureLabel := "feature"
	otherColor := "112233"
	_, _, err := githubClient.Issues.CreateLabel(context.Background(), suite.testOrg, suite.testRepo, &github.Label{Name: &featureLabel, Color: &otherColor})
	suite.NoError(err)
	suite.runUnifyLabelCommand(true)
	label, _, err := githubClient.Issues.GetLabel(context.Background(), suite.testOrg, suite.testRepo, "feature")
	suite.NoError(err)
	suite.Equal(*label.Name, "feature")
	suite.Equal(*label.Color, "88ee66")
}

func (suite *UnifyLabelsSuite) TestDeleteLabel() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	unknownLabel := "unknown123"
	_, _, err := githubClient.Issues.CreateLabel(context.Background(), suite.testOrg, suite.testRepo, &github.Label{Name: &unknownLabel})
	suite.NoError(err)
	suite.runUnifyLabelCommand(true)
	labels, _, err := githubClient.Issues.ListLabels(context.Background(), suite.testOrg, suite.testRepo, &github.ListOptions{PerPage: 100})
	suite.NoError(err)
	labelNames := []string{}
	for _, label := range labels {
		labelNames = append(labelNames, *label.Name)
	}
	suite.Assert().NotContains(labelNames, unknownLabel)
}

func (suite *UnifyLabelsSuite) cleanup() {
	suite.deleteAllLabels()
	suite.closeAllIssues()
}

func (suite *UnifyLabelsSuite) deleteAllLabels() {
	githubClient := getGithubClient()
	labels, _, err := githubClient.Issues.ListLabels(context.Background(), suite.testOrg, suite.testRepo, &github.ListOptions{PerPage: 100})
	suite.NoError(err)
	for _, label := range labels {
		_, err = githubClient.Issues.DeleteLabel(context.Background(), suite.testOrg, suite.testRepo, *label.Name)
		suite.NoError(err)
	}
}

func (suite *UnifyLabelsSuite) closeAllIssues() {
	githubClient := getGithubClient()
	issues, _, err := githubClient.Issues.ListByRepo(context.Background(), suite.testOrg, suite.testRepo, &github.IssueListByRepoOptions{ListOptions: github.ListOptions{PerPage: 100}})
	suite.NoError(err)
	for _, issue := range issues {
		closed := "closed"
		_, _, err = githubClient.Issues.Edit(context.Background(), suite.testOrg, suite.testRepo, *issue.Number, &github.IssueRequest{State: &closed})
		suite.NoError(err)
	}
}

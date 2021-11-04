package cmd

import (
	"context"
	"testing"

	"github.com/google/go-github/v39/github"
	"github.com/stretchr/testify/suite"
)

type UnifyLablesSuite struct {
	suite.Suite
}

func TestUnifyLablesSuite(t *testing.T) {
	suite.Run(t, new(UnifyLablesSuite))
}

func (suite *UnifyLablesSuite) TestCreateLables() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	unifyLabels(testRepo, githubClient, true) // should create labels
	label, _, err := githubClient.Issues.GetLabel(context.Background(), testOrg, testRepo, "feature")
	suite.NoError(err)
	suite.Equal(*label.Name, "feature")
	suite.Equal(*label.Color, "88ee66")
}

func (suite *UnifyLablesSuite) TestRenameLabel() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	labelName := "blocked"
	issueName := "TestIssue"
	issue, _, err := githubClient.Issues.Create(context.Background(), testOrg, testRepo, &github.IssueRequest{Title: &issueName, Labels: &[]string{labelName}})
	suite.NoError(err)
	unifyLabels(testRepo, githubClient, true) // should update label to blocked:yes
	updatedIssue, _, err := githubClient.Issues.Get(context.Background(), testOrg, testRepo, *issue.Number)
	suite.NoError(err)
	var labelNames []string
	for _, label := range updatedIssue.Labels {
		labelNames = append(labelNames, *label.Name)
	}
	suite.Assert().Contains(labelNames, "blocked:yes")
}

func (suite *UnifyLablesSuite) TestMigrateLabel() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	blockedLabel := "blocked"
	blockedYesLabel := "blocked:yes"
	_, _, err := githubClient.Issues.CreateLabel(context.Background(), testOrg, testRepo, &github.Label{Name: &blockedYesLabel})
	suite.NoError(err)
	issueName := "TestIssue"
	issue, _, err := githubClient.Issues.Create(context.Background(), testOrg, testRepo, &github.IssueRequest{Title: &issueName, Labels: &[]string{blockedLabel}})
	suite.NoError(err)
	unifyLabels(testRepo, githubClient, true) // should update label to blocked:yes
	updatedIssue, _, err := githubClient.Issues.Get(context.Background(), testOrg, testRepo, *issue.Number)
	suite.NoError(err)
	var labelNames []string
	for _, label := range updatedIssue.Labels {
		labelNames = append(labelNames, *label.Name)
	}
	suite.Assert().Contains(labelNames, blockedYesLabel)
	suite.Assert().NotContains(labelNames, blockedLabel)
}

func (suite *UnifyLablesSuite) TestChangeColor() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	featureLabel := "feature"
	otherColor := "112233"
	_, _, err := githubClient.Issues.CreateLabel(context.Background(), testOrg, testRepo, &github.Label{Name: &featureLabel, Color: &otherColor})
	suite.NoError(err)
	unifyLabels(testRepo, githubClient, true)
	label, _, err := githubClient.Issues.GetLabel(context.Background(), testOrg, testRepo, "feature")
	suite.NoError(err)
	suite.Equal(*label.Name, "feature")
	suite.Equal(*label.Color, "88ee66")
}

func (suite *UnifyLablesSuite) TestDeleteLabel() {
	suite.cleanup()       // to be sure we start with a defined state
	defer suite.cleanup() // to leave a clean repo
	githubClient := getGithubClient()
	unknownLabel := "unknown123"
	_, _, err := githubClient.Issues.CreateLabel(context.Background(), testOrg, testRepo, &github.Label{Name: &unknownLabel})
	suite.NoError(err)
	unifyLabels(testRepo, githubClient, true) // should create labels
	labels, _, err := githubClient.Issues.ListLabels(context.Background(), testOrg, testRepo, &github.ListOptions{PerPage: 100})
	suite.NoError(err)
	labelNames := []string{}
	for _, label := range labels {
		labelNames = append(labelNames, *label.Name)
	}
	suite.Assert().NotContains(labelNames, unknownLabel)
}

func (suite *UnifyLablesSuite) cleanup() {
	suite.deleteAllLabels()
	suite.closeAllIssues()
}

func (suite *UnifyLablesSuite) deleteAllLabels() {
	githubClient := getGithubClient()
	labels, _, err := githubClient.Issues.ListLabels(context.Background(), testOrg, testRepo, &github.ListOptions{PerPage: 100})
	suite.NoError(err)
	for _, label := range labels {
		_, err = githubClient.Issues.DeleteLabel(context.Background(), testOrg, testRepo, *label.Name)
		suite.NoError(err)
	}
}

func (suite *UnifyLablesSuite) closeAllIssues() {
	githubClient := getGithubClient()
	issues, _, err := githubClient.Issues.ListByRepo(context.Background(), testOrg, testRepo, &github.IssueListByRepoOptions{ListOptions: github.ListOptions{PerPage: 100}})
	suite.NoError(err)
	for _, issue := range issues {
		closed := "closed"
		_, _, err = githubClient.Issues.Edit(context.Background(), testOrg, testRepo, *issue.Number, &github.IssueRequest{State: &closed})
		suite.NoError(err)
	}
}

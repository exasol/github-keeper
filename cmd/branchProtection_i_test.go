package cmd

import (
	"context"
	"fmt"
	"github.com/google/go-github/v39/github"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BranchProtectionSuite struct {
	IntegrationTestSuite
}

func TestBranchProtectionSuite(t *testing.T) {
	suite.Run(t, new(BranchProtectionSuite))
}

func (suite *BranchProtectionSuite) TestCreateBranchProtection() {
	suite.cleanup()
	defer suite.cleanup()
	verifier := BranchProtectionVerifier{repoName: suite.testRepo, client: getGithubClient()}
	verifier.CheckIfBranchProtectionIsApplied(true)
	protection, _, err := suite.githubClient.Repositories.GetBranchProtection(context.Background(), suite.testOrg, suite.testRepo, "master")
	suite.NoError(err)
	suite.assertBranchProtection(protection)
}

func (suite *BranchProtectionSuite) assertBranchProtection(protection *github.Protection) {
	suite.Assert().False(protection.AllowForcePushes.Enabled)
	suite.Assert().True(protection.RequiredPullRequestReviews.DismissStaleReviews)
	suite.Assert().True(protection.EnforceAdmins.Enabled)
	suite.Assert().True(protection.RequiredPullRequestReviews.RequireCodeOwnerReviews)
	suite.Assert().Equal(protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, 1)
	suite.Assert().True(protection.RequiredStatusChecks.Strict)
	suite.Assert().Contains(protection.RequiredStatusChecks.Contexts, "linkChecker")
	suite.Assert().Contains(protection.RequiredStatusChecks.Contexts, "SonarCloud Code Analysis")
}

func (suite *BranchProtectionSuite) TestBranchProtectionMissing() {
	suite.cleanup()
	defer suite.cleanup()
	output := suite.CaptureOutput(func() {
		verifier := BranchProtectionVerifier{repoName: suite.testRepo, client: getGithubClient()}
		verifier.CheckIfBranchProtectionIsApplied(false)
	})
	suite.Assert().Equal("exasol/testing-release-robot does not have a branch protection rule for default branch master. Use --fix to create it. This error can also happen if you don't have admin privileges on the repo.", output)
}

func (suite *BranchProtectionSuite) TestUpdateIncompleteBranchProtection() {
	suite.cleanup()
	defer suite.cleanup()
	suite.createEmptyBranchProtection()
	verifier := BranchProtectionVerifier{repoName: suite.testRepo, client: getGithubClient()}
	verifier.CheckIfBranchProtectionIsApplied(true)
	protection, _, err := suite.githubClient.Repositories.GetBranchProtection(context.Background(), suite.testOrg, suite.testRepo, "master")
	suite.NoError(err)
	suite.assertBranchProtection(protection)
}

func (suite *BranchProtectionSuite) TestBranchProtectionUpdatePreserversExistingChecks() {
	suite.cleanup()
	defer suite.cleanup()
	request := github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Contexts: []string{"myAdditionalCheck"},
		},
	}
	_, _, err := suite.githubClient.Repositories.UpdateBranchProtection(context.Background(), suite.testOrg, suite.testRepo, suite.testDefaultBranch, &request)
	suite.NoError(err)
	verifier := BranchProtectionVerifier{repoName: suite.testRepo, client: getGithubClient()}
	verifier.CheckIfBranchProtectionIsApplied(true)
	protection, _, err := suite.githubClient.Repositories.GetBranchProtection(context.Background(), suite.testOrg, suite.testRepo, "master")
	suite.NoError(err)
	suite.Contains(protection.RequiredStatusChecks.Contexts, "myAdditionalCheck")
}

func (suite *BranchProtectionSuite) TestBranchProtectionIncomplete() {
	suite.cleanup()
	defer suite.cleanup()
	suite.createEmptyBranchProtection()
	output := suite.CaptureOutput(func() {
		verifier := BranchProtectionVerifier{repoName: suite.testRepo, client: getGithubClient()}
		verifier.CheckIfBranchProtectionIsApplied(false)
	})
	suite.Assert().Equal("exasol/testing-release-robot has a branch protection for default branch master that is not compliant to our standards. Use --fix to update.\n", output)
}

func (suite *BranchProtectionSuite) createEmptyBranchProtection() {
	request := github.ProtectionRequest{}
	_, _, err := suite.githubClient.Repositories.UpdateBranchProtection(context.Background(), suite.testOrg, suite.testRepo, suite.testDefaultBranch, &request)
	suite.NoError(err)
}

func (suite *BranchProtectionSuite) cleanup() {
	_, err := suite.githubClient.Repositories.RemoveBranchProtection(context.Background(), suite.testOrg, suite.testRepo, "master")
	if err != nil {
		if strings.Contains(err.Error(), "Branch not protected") {
			//ignore
		} else {
			suite.NoError(err)
		}
	}
}

func (suite *BranchProtectionSuite) TestGetChecksForWorkflowContentWithListSyntax() {
	verifier := BranchProtectionVerifier{}
	definition, err := verifier.parseWorkflowDefinition(`
name: CI Build
on:
  - push
jobs:
  build:
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	suite.Contains(definition.JobsNames, "build")
	suite.Contains(definition.Trigger, "push")
	suite.Contains(definition.Name, "CI Build")

}

func (suite *BranchProtectionSuite) TestGetChecksForWorkflowContentWithMapSyntax() {
	verifier := BranchProtectionVerifier{}
	definition, err := verifier.parseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	suite.Contains(definition.JobsNames, "build")
	suite.Contains(definition.Trigger, "push")
	suite.Contains(definition.Name, "CI Build")

}

type TestHasWorkflowPushOrPrTriggerCase struct {
	trigger        []string
	expectedResult bool
}

func (suite *BranchProtectionSuite) TestHasWorkflowPushOrPrTrigger() {
	cases := []TestHasWorkflowPushOrPrTriggerCase{
		{trigger: []string{""}, expectedResult: false},
		{trigger: []string{"other"}, expectedResult: false},
		{trigger: []string{"push"}, expectedResult: true},
		{trigger: []string{"pull_request"}, expectedResult: true},
		{trigger: []string{"other", "push"}, expectedResult: true},
	}
	verifier := BranchProtectionVerifier{}
	for _, testCase := range cases {
		suite.Run(fmt.Sprintf("trigger: %v", testCase.trigger), func() {
			definition := workflowDefinition{Trigger: testCase.trigger}
			result := verifier.hasWorkflowPushOrPrTrigger(&definition)
			suite.Equal(testCase.expectedResult, result)
		})
	}
}

func (suite *BranchProtectionSuite) TestCheckIfBranchRestrictionsAreAppliedWithEqualInputs() {
	verifier := BranchProtectionVerifier{}
	testUserName := "testUser"
	testTeamName := "testGroup"
	testAppName := "testApp"
	existing := github.BranchRestrictions{Users: []*github.User{{Name: &testUserName}}, Teams: []*github.Team{{Name: &testTeamName}}, Apps: []*github.App{{Name: &testAppName}}}
	request := github.BranchRestrictionsRequest{Users: []string{testUserName}, Teams: []string{testTeamName}, Apps: []string{testAppName}}
	suite.Assert().True(verifier.checkIfBranchRestrictionsAreApplied(&existing, &request))
}

func (suite *BranchProtectionSuite) TestCheckIfBranchRestrictionsAreAppliedWithNonEqualUserName() {
	verifier := BranchProtectionVerifier{}
	testUserName := "testUser"
	testTeamName := "testGroup"
	testAppName := "testApp"
	existing := github.BranchRestrictions{Users: []*github.User{{Name: &testUserName}}, Teams: []*github.Team{{Name: &testTeamName}}, Apps: []*github.App{{Name: &testAppName}}}
	request := github.BranchRestrictionsRequest{Users: []string{"otherUser"}, Teams: []string{testTeamName}, Apps: []string{testAppName}}
	suite.Assert().False(verifier.checkIfBranchRestrictionsAreApplied(&existing, &request))
}

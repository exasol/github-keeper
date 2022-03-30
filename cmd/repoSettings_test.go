package cmd

import (
	"context"
	"testing"

	"github.com/google/go-github/v43/github"
	"github.com/stretchr/testify/suite"
)

type RepoSettingsSuite struct {
	IntegrationTestSuite
}

func TestRepoSettingsSuite(t *testing.T) {
	suite.Run(t, new(RepoSettingsSuite))
}

func (suite RepoSettingsSuite) TestInvalidSettings() {
	suite.disableAutoMergeAndBranchDeleting()
	verifier := RepoSettingsVerifier{repo: suite.testRepo, org: suite.testOrg, githubClient: suite.githubClient}
	output := suite.CaptureOutput(func() {
		verifier.VerifyRepoSettings(false)
	})
	suite.Equal(output, "The repository testing-release-robot has outdated repo settings.\n")
}

func (suite RepoSettingsSuite) TestFix() {
	suite.disableAutoMergeAndBranchDeleting()
	verifier := RepoSettingsVerifier{repo: suite.testRepo, org: suite.testOrg, githubClient: suite.githubClient}
	verifier.VerifyRepoSettings(true)
	repo, _, err := suite.githubClient.Repositories.Get(context.Background(), suite.testOrg, suite.testRepo)
	suite.NoError(err)
	suite.Assert().True(*repo.AllowAutoMerge)
	suite.Assert().True(*repo.DeleteBranchOnMerge)
}

func (suite RepoSettingsSuite) TestSettingsValidAfterFix() {
	suite.disableAutoMergeAndBranchDeleting()
	verifier := RepoSettingsVerifier{repo: suite.testRepo, org: suite.testOrg, githubClient: suite.githubClient}
	verifier.VerifyRepoSettings(true)
	output := suite.CaptureOutput(func() {
		verifier.VerifyRepoSettings(false)
	})
	suite.Equal(output, "")
}

func (suite RepoSettingsSuite) disableAutoMergeAndBranchDeleting() {
	falsePointer := false
	_, _, err := suite.githubClient.Repositories.Edit(context.Background(), suite.testOrg, suite.testRepo, &github.Repository{AllowAutoMerge: &falsePointer, DeleteBranchOnMerge: &falsePointer})
	suite.NoError(err)
}

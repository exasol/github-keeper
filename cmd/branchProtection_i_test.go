package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BranchProtectionSuite struct {
	suite.Suite
}

func TestBranchProtectionSuite(t *testing.T) {
	suite.Run(t, new(BranchProtectionSuite))
}

func (suite *BranchProtectionSuite) TestCreateBranchProtection() {
	suite.cleanup()
	defer suite.cleanup()
	githubClient := getGithubClient()
	createBranchProtection(testRepo, githubClient)
	protection, _, err := githubClient.Repositories.GetBranchProtection(context.Background(), testOrg, testRepo, "master")
	onError(err)
	suite.Assert().False(protection.AllowForcePushes.Enabled)
	suite.Assert().True(protection.RequiredPullRequestReviews.DismissStaleReviews)
	suite.Assert().True(protection.RequiredPullRequestReviews.RequireCodeOwnerReviews)
	suite.Assert().Equal(protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, 1)
	suite.Assert().True(protection.RequiredStatusChecks.Strict)
}

func (suite *BranchProtectionSuite) cleanup() {
	client := getGithubClient()
	_, err := client.Repositories.RemoveBranchProtection(context.Background(), testOrg, testRepo, "master")
	if err != nil {
		if strings.Contains(err.Error(), "Branch not protected") {
			//ignore
		} else {
			onError(err)
		}
	}
}

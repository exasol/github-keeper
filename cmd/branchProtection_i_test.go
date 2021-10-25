package cmd

import (
	"bytes"
	"context"
	"github.com/google/go-github/v39/github"
	"io"
	"log"
	"os"
	"strings"
	"sync"
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
	err := branchProtectionCmd.Flags().Set("fix", "true")
	suite.NoError(err)
	branchProtectionCmd.Run(branchProtectionCmd, []string{testRepo})
	protection, _, err := githubClient.Repositories.GetBranchProtection(context.Background(), testOrg, testRepo, "master")
	suite.NoError(err)
	suite.assertBranchProtection(protection)
}

func (suite *BranchProtectionSuite) assertBranchProtection(protection *github.Protection) {
	suite.Assert().False(protection.AllowForcePushes.Enabled)
	suite.Assert().True(protection.RequiredPullRequestReviews.DismissStaleReviews)
	suite.Assert().True(protection.RequiredPullRequestReviews.RequireCodeOwnerReviews)
	suite.Assert().Equal(protection.RequiredPullRequestReviews.RequiredApprovingReviewCount, 1)
	suite.Assert().True(protection.RequiredStatusChecks.Strict)
	suite.Assert().Contains(protection.RequiredStatusChecks.Contexts, "linkChecker")
}

func (suite *BranchProtectionSuite) TestBranchProtectionMissing() {
	suite.cleanup()
	defer suite.cleanup()
	output := captureOutput(func() {
		branchProtectionCmd.Run(branchProtectionCmd, []string{testRepo})
	})
	suite.Assert().Equal("exasol/testing-release-robot does not have a branch protection rule for default branch master. Use --fix to create it. This error can also happen if you don't have admin privileges on the repo.", output)
}

func (suite *BranchProtectionSuite) TestUpdateIncompleteBranchProtection() {
	suite.cleanup()
	defer suite.cleanup()
	suite.createEmptyBranchProtection()
	err := branchProtectionCmd.Flags().Set("fix", "true")
	suite.NoError(err)
	branchProtectionCmd.Run(branchProtectionCmd, []string{testRepo})
	githubClient := getGithubClient()
	protection, _, err := githubClient.Repositories.GetBranchProtection(context.Background(), testOrg, testRepo, "master")
	suite.NoError(err)
	suite.assertBranchProtection(protection)
}

func (suite *BranchProtectionSuite) TestBranchProtectionUpdatePreserversExistingChecks() {
	suite.cleanup()
	defer suite.cleanup()
	githubClient := getGithubClient()
	request := github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Contexts: []string{"myAdditionalCheck"},
		},
	}
	_, _, err := githubClient.Repositories.UpdateBranchProtection(context.Background(), testOrg, testRepo, testDefaultBranch, &request)
	suite.NoError(err)
	err = branchProtectionCmd.Flags().Set("fix", "true")
	suite.NoError(err)
	branchProtectionCmd.Run(branchProtectionCmd, []string{testRepo})
	protection, _, err := githubClient.Repositories.GetBranchProtection(context.Background(), testOrg, testRepo, "master")
	suite.NoError(err)
	suite.Contains(protection.RequiredStatusChecks.Contexts, "myAdditionalCheck")
}

func (suite *BranchProtectionSuite) TestBranchProtectionIncomplete() {
	suite.cleanup()
	defer suite.cleanup()
	suite.createEmptyBranchProtection()
	output := captureOutput(func() {
		branchProtectionCmd.Run(branchProtectionCmd, []string{testRepo})
	})
	suite.Assert().Equal("exasol/testing-release-robot has a branch protection for default branch master that is not compliant to our standards. Use --fix to update.\n", output)
}

func (suite *BranchProtectionSuite) createEmptyBranchProtection() {
	client := getGithubClient()
	request := github.ProtectionRequest{}
	_, _, err := client.Repositories.UpdateBranchProtection(context.Background(), testOrg, testRepo, testDefaultBranch, &request)
	suite.NoError(err)
}

func captureOutput(functionToCapture func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	isReaderReady := new(sync.WaitGroup)
	isReaderReady.Add(1)
	go func() {
		var buffer bytes.Buffer
		isReaderReady.Done()
		_, err := io.Copy(&buffer, reader) //blocking
		if err != nil {
			panic(err)
		}
		out <- buffer.String()
	}()
	isReaderReady.Wait()
	functionToCapture()
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	return <-out
}

func (suite *BranchProtectionSuite) cleanup() {
	client := getGithubClient()
	_, err := client.Repositories.RemoveBranchProtection(context.Background(), testOrg, testRepo, "master")
	if err != nil {
		if strings.Contains(err.Error(), "Branch not protected") {
			//ignore
		} else {
			suite.NoError(err)
		}
	}
}

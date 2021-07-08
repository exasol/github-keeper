package cmd

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

var branchProtectionCmd = &cobra.Command{
	Use:   "create-branch-protection <repo-name>",
	Args:  cobra.MinimumNArgs(1),
	Short: "Setup a branch protection for a given repo",
	Run: func(cmd *cobra.Command, args []string) {
		client := getGithubClient()
		for _, repo := range args {
			createBranchProtection(repo, client)
		}

	},
}

func createBranchProtection(repoName string, client *github.Client) {
	repo, _, err := client.Repositories.Get(context.Background(), "exasol", repoName)
	if err != nil {
		panic(fmt.Sprintf("Failed to get repository exasol/%v. Cause: %v", repoName, err.Error()))
	}
	branch := *repo.DefaultBranch
	protectionRequest := createProtectionRequest()
	_, _, err = client.Repositories.UpdateBranchProtection(context.Background(), "exasol", repoName, branch, &protectionRequest)
	if err != nil {
		panic(fmt.Sprintf("Failed to create branch protection for exasol/%v/%v. Cause: %v", repoName, branch, err.Error()))
	} else {
		fmt.Printf("Sucessfully updated branch protection for %v.\n", repoName)
	}
}

func createProtectionRequest() github.ProtectionRequest {
	return github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{},
		},
		RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
			DismissStaleReviews:          true,
			RequireCodeOwnerReviews:      true,
			RequiredApprovingReviewCount: 1,
		},
		EnforceAdmins: false,
		Restrictions:  nil,
	}
}

func init() {
	rootCmd.AddCommand(branchProtectionCmd)
}

package cmd

import (
	"context"
	"fmt"
	"github.com/google/go-github/v39/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var branchProtectionCmd = &cobra.Command{
	Use:   "create-branch-protection <repo-name>",
	Args:  cobra.MinimumNArgs(1),
	Short: "Setup a branch protection for a given repo",
	Run: func(cmd *cobra.Command, args []string) {
		client := getGithubClient()
		for _, repo := range args {
			verifier := BranchProtectionVerifier{client: client, repoName: repo}
			verifier.createBranchProtection()
		}
	},
}

type BranchProtectionVerifier struct {
	repoName string
	client   *github.Client
}

func (verifier BranchProtectionVerifier) createBranchProtection() {
	branch := verifier.getDefaultBranch()
	protectionRequest := verifier.createProtectionRequest()
	_, _, err := verifier.client.Repositories.UpdateBranchProtection(context.Background(), "exasol", verifier.repoName, branch, &protectionRequest)
	if err != nil {
		panic(fmt.Sprintf("Failed to create branch protection for exasol/%v/%v. Cause: %v", verifier.repoName, branch, err.Error()))
	} else {
		fmt.Printf("Sucessfully updated branch protection for %v.\n", verifier.repoName)
	}
}

func (verifier BranchProtectionVerifier) getDefaultBranch() string {
	repo, _, err := verifier.client.Repositories.Get(context.Background(), "exasol", verifier.repoName)
	if err != nil {
		panic(fmt.Sprintf("Failed to get repository exasol/%v. Cause: %v", verifier.repoName, err.Error()))
	}
	branch := *repo.DefaultBranch
	return branch
}

func (verifier BranchProtectionVerifier) createProtectionRequest() github.ProtectionRequest {
	allowForcePushes := false
	requiredChecks, err := verifier.getRequiredChecks()
	if err != nil {
		panic(fmt.Sprintf("Failed to get required checks for repository %v. Cause: %v", verifier.repoName, err.Error()))
	}
	return github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Strict:   true,
			Contexts: requiredChecks,
		},
		RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
			DismissStaleReviews:          true,
			RequireCodeOwnerReviews:      true,
			RequiredApprovingReviewCount: 1,
		},
		EnforceAdmins:    false,
		Restrictions:     nil,
		AllowForcePushes: &allowForcePushes,
	}
}

func (verifier BranchProtectionVerifier) getRequiredChecks() (result []string, err error) {
	_, directory, _, err := verifier.client.Repositories.GetContents(context.Background(), "exasol", verifier.repoName, ".github/workflows/", &github.RepositoryContentGetOptions{})
	if err != nil {
		return nil, err
	}
	for _, fileDesc := range directory {
		workflowFilePath := fileDesc.Path
		requiredChecksForWorkflow, err := verifier.getChecksForWorkflow(workflowFilePath)
		if err != nil {
			return nil, err
		}
		result = append(result, requiredChecksForWorkflow...)
	}
	return result, err
}

func (verifier BranchProtectionVerifier) getChecksForWorkflow(workflowFilePath *string) ([]string, error) {
	var result []string
	content, err := verifier.downloadWorkflowFile(*workflowFilePath)
	if err != nil {
		return nil, err
	}
	parsedYaml, err := verifier.parseWorkflowDefinition(content)
	if err != nil {
		return nil, err
	}
	_, foundPush := parsedYaml.On["push"]
	_, foundPullRequest := parsedYaml.On["pull_request"]
	if foundPush || foundPullRequest {
		for jobName, _ := range parsedYaml.Jobs {
			result = append(result, jobName)
		}
	}
	return result, nil
}

func (verifier BranchProtectionVerifier) downloadWorkflowFile(path string) (string, error) {
	workflowFile, _, _, err := verifier.client.Repositories.GetContents(context.Background(), "exasol", verifier.repoName, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		return "", err
	}
	return workflowFile.GetContent()
}

func (verifier BranchProtectionVerifier) parseWorkflowDefinition(content string) (workflowDefinition, error) {
	parsedYaml := workflowDefinition{}
	err := yaml.Unmarshal([]byte(content), &parsedYaml)
	if err != nil {
		return parsedYaml, err
	}
	return parsedYaml, nil
}

type workflowDefinition struct {
	Name string                 `yaml:"name"`
	On   map[string]interface{} `yaml:"on"`
	Jobs map[string]interface{} `yaml:"jobs"`
}

func init() {
	rootCmd.AddCommand(branchProtectionCmd)
}

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
		fix, err := cmd.Flags().GetBool("fix")
		if err != nil {
			panic("Could not read parameter fix")
		}
		for _, repo := range args {
			verifier := BranchProtectionVerifier{client: client, repoName: repo}
			verifier.verifyBranchProtection(fix)
		}
	},
}

type BranchProtectionVerifier struct {
	repoName string
	client   *github.Client
}

type BranchProtectionProblemHandler interface {
	createBranchProtection(repo string, branch string, protection *github.ProtectionRequest)
	updateProtection(repo string, branch string, protection *github.ProtectionRequest)
}

type LogBranchProtectionProblemHandler struct {
}

func (logHandler LogBranchProtectionProblemHandler) createBranchProtection(repo string, branch string, protection *github.ProtectionRequest) {
	fmt.Printf("exasol/%v does not have a branch protection rule for default branch %v. Use --fix to create it. This error can also happen if you don't have admin privileges on the repo.", repo, branch)
}

type FixBranchProtectionProblemHandler struct {
	client *github.Client
}

func (logHandler LogBranchProtectionProblemHandler) updateProtection(repo string, branch string, protection *github.ProtectionRequest) {
	fmt.Printf("exasol/%v has a branch protection for default branch %v that is not compliant to our standards. Use --fix to update.\n", repo, branch)
}

func (handler FixBranchProtectionProblemHandler) createBranchProtection(repo string, branch string, protection *github.ProtectionRequest) {
	_, _, err := handler.client.Repositories.UpdateBranchProtection(context.Background(), "exasol", repo, branch, protection)
	if err != nil {
		panic(fmt.Sprintf("Failed to create branch protection for exasol/%v/%v. Cause: %v", repo, branch, err.Error()))
	} else {
		fmt.Printf("Sucessfully created branch protection for exasol/%v/%v.\n", repo, branch)
	}
}

func (handler FixBranchProtectionProblemHandler) updateProtection(repo string, branch string, protection *github.ProtectionRequest) {
	handler.createBranchProtection(repo, branch, protection)
}

func (verifier BranchProtectionVerifier) verifyBranchProtection(fix bool) {
	problemHandler := verifier.getProblemHandler(fix)
	branch := verifier.getDefaultBranch()
	existingProtection, resp, _ := verifier.client.Repositories.GetBranchProtection(context.Background(), "exasol", verifier.repoName, branch)
	protectionRequest := verifier.createProtectionRequest()
	if resp.StatusCode == 404 {
		problemHandler.createBranchProtection(verifier.repoName, branch, &protectionRequest)
	} else {
		if !(existingProtection.AllowForcePushes.Enabled == *protectionRequest.AllowForcePushes &&
			checkReviewCompliance(existingProtection.RequiredPullRequestReviews, protectionRequest.RequiredPullRequestReviews) &&
			verifier.checkStatusCheckCompliance(existingProtection.RequiredStatusChecks, protectionRequest.RequiredStatusChecks)) {
			verifier.addExistingChecksToRequest(existingProtection, protectionRequest)
			problemHandler.updateProtection(verifier.repoName, branch, &protectionRequest)
		}
	}
}

func (verifier BranchProtectionVerifier) addExistingChecksToRequest(existingProtection *github.Protection, protectionRequest github.ProtectionRequest) {
	if existingProtection == nil || existingProtection.RequiredStatusChecks == nil || existingProtection.RequiredStatusChecks.Contexts == nil {
		return
	}
	for _, existingCheck := range existingProtection.RequiredStatusChecks.Contexts {
		if !verifier.containsValue(existingCheck, protectionRequest.RequiredStatusChecks.Contexts) {
			protectionRequest.RequiredStatusChecks.Contexts = append(protectionRequest.RequiredStatusChecks.Contexts, existingCheck)
		}
	}
}

func (verifier BranchProtectionVerifier) checkStatusCheckCompliance(existing *github.RequiredStatusChecks, request *github.RequiredStatusChecks) bool {
	if existing == nil || request == nil {
		return false
	}
	for _, requiredCheck := range request.Contexts {
		if !verifier.containsValue(requiredCheck, existing.Contexts) {
			return false
		}
	}
	return existing.Strict == request.Strict
}

func (verifier BranchProtectionVerifier) containsValue(value string, values []string) bool {
	for _, existingCheck := range values {
		if existingCheck == value {
			return true
		}
	}
	return false
}

func checkReviewCompliance(existing *github.PullRequestReviewsEnforcement, request *github.PullRequestReviewsEnforcementRequest) bool {
	return existing != nil && request != nil &&
		existing.RequiredApprovingReviewCount >= request.RequiredApprovingReviewCount &&
		existing.DismissStaleReviews == request.DismissStaleReviews &&
		existing.RequireCodeOwnerReviews == request.RequireCodeOwnerReviews
}

func (verifier BranchProtectionVerifier) getProblemHandler(fix bool) BranchProtectionProblemHandler {
	var problemHandler BranchProtectionProblemHandler
	if fix {
		problemHandler = FixBranchProtectionProblemHandler{verifier.client}
	} else {
		problemHandler = LogBranchProtectionProblemHandler{}
	}
	return problemHandler
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
	content, err := verifier.downloadFile(*workflowFilePath)
	if err != nil {
		return nil, err
	}
	return verifier.getChecksForWorkflowContent(content)
}

func (verifier BranchProtectionVerifier) getChecksForWorkflowContent(content string) ([]string, error) {
	var result []string
	workflow, err := verifier.parseWorkflowDefinition(content)
	if err != nil {
		return nil, err
	}
	hasWorkflowPushOrPrTrigger := verifier.hasWorkflowPushOrPrTrigger(workflow)
	if hasWorkflowPushOrPrTrigger {
		return workflow.JobsNames, nil
	}
	return result, nil
}

func (verifier BranchProtectionVerifier) hasWorkflowPushOrPrTrigger(parsedYaml *workflowDefinition) bool {
	for _, trigger := range parsedYaml.Trigger {
		if trigger == "push" || trigger == "pull_request" {
			return true
		}
	}
	return false
}

func (verifier BranchProtectionVerifier) downloadFile(path string) (string, error) {
	workflowFile, _, _, err := verifier.client.Repositories.GetContents(context.Background(), "exasol", verifier.repoName, path, &github.RepositoryContentGetOptions{})
	if err != nil {
		return "", err
	}
	return workflowFile.GetContent()
}

func (verifier BranchProtectionVerifier) parseWorkflowDefinition(content string) (*workflowDefinition, error) {
	parsedYaml := workflowDefinitionInt{}
	err := yaml.Unmarshal([]byte(content), &parsedYaml)
	if err != nil {
		return nil, err
	}
	trigger, err := getTriggersOfWorkflowDefinition(&parsedYaml)
	if err != nil {
		return nil, err
	}
	var jobNames []string
	for jobName := range parsedYaml.Jobs {
		jobNames = append(jobNames, jobName)
	}
	definition := workflowDefinition{Name: parsedYaml.Name, JobsNames: jobNames, Trigger: trigger}
	return &definition, nil
}

func getTriggersOfWorkflowDefinition(parsedYaml *workflowDefinitionInt) ([]string, error) {
	if triggerMap, hasTriggerMap := parsedYaml.On.(map[interface{}]interface{}); hasTriggerMap {
		var result []string
		for trigger := range triggerMap {
			result = append(result, trigger.(string))
		}
		return result, nil
	} else if triggerList, hasTriggerList := parsedYaml.On.([]interface{}); hasTriggerList {
		var result []string
		for _, trigger := range triggerList {
			result = append(result, trigger.(string))
		}
		return result, nil
	} else {
		return nil, fmt.Errorf("the GitHub workflow '%v' has a unimplemented trigger definition style", parsedYaml.Name)
	}
}

type workflowDefinitionInt struct {
	Name string                 `yaml:"name"`
	On   interface{}            `yaml:"on"`
	Jobs map[string]interface{} `yaml:"jobs"`
}

type workflowDefinition struct {
	Name      string
	Trigger   []string
	JobsNames []string
}

func init() {
	branchProtectionCmd.Flags().Bool("fix", false, "If this flag is set, github-keeper creates the branch protection. Otherwise it just prints the diff.")
	rootCmd.AddCommand(branchProtectionCmd)
}

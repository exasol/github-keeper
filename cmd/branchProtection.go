package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v39/github"
)

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

func (verifier BranchProtectionVerifier) CheckIfBranchProtectionIsApplied(fix bool) {
	problemHandler := verifier.getProblemHandler(fix)
	repo := verifier.getRepo()
	branch := *repo.DefaultBranch
	existingProtection, resp, _ := verifier.client.Repositories.GetBranchProtection(context.Background(), "exasol", verifier.repoName, branch)
	protectionRequest := verifier.createProtectionRequest(verifier.isSonarRequired(repo.Language))
	if resp.StatusCode == 404 {
		problemHandler.createBranchProtection(verifier.repoName, branch, &protectionRequest)
	} else {
		if !(existingProtection.AllowForcePushes.Enabled == *protectionRequest.AllowForcePushes &&
			existingProtection.EnforceAdmins.Enabled == protectionRequest.EnforceAdmins &&
			verifier.checkIfPrReviewPolicyIsApplied(existingProtection.RequiredPullRequestReviews, protectionRequest.RequiredPullRequestReviews) &&
			verifier.checkIfStatusCheckPolicyIsApplied(existingProtection.RequiredStatusChecks, protectionRequest.RequiredStatusChecks) &&
			verifier.checkIfBranchRestrictionsAreApplied(existingProtection.Restrictions, protectionRequest.Restrictions)) {
			verifier.addExistingChecksToRequest(existingProtection, protectionRequest)
			problemHandler.updateProtection(verifier.repoName, branch, &protectionRequest)
		}
	}
}

func (verifier BranchProtectionVerifier) isSonarRequired(language *string) bool {
	return language != nil && (*language == "Scala" || *language == "Java" || *language == "Go")
}

func (verifier BranchProtectionVerifier) getRepo() *github.Repository {
	repo, _, err := verifier.client.Repositories.Get(context.Background(), "exasol", verifier.repoName)
	if err != nil {
		panic(fmt.Sprintf("Failed to get repository exasol/%v. Cause: %v", verifier.repoName, err.Error()))
	}
	return repo
}

func (verifier BranchProtectionVerifier) addExistingChecksToRequest(existingProtection *github.Protection, protectionRequest github.ProtectionRequest) {
	if existingProtection == nil || existingProtection.RequiredStatusChecks == nil || existingProtection.RequiredStatusChecks.Contexts == nil {
		return
	}
	for _, existingCheck := range existingProtection.RequiredStatusChecks.Contexts {
		if !verifier.containsValue(protectionRequest.RequiredStatusChecks.Contexts, existingCheck) {
			protectionRequest.RequiredStatusChecks.Contexts = append(protectionRequest.RequiredStatusChecks.Contexts, existingCheck)
		}
	}
}

func (verifier BranchProtectionVerifier) checkIfStatusCheckPolicyIsApplied(existing *github.RequiredStatusChecks, request *github.RequiredStatusChecks) bool {
	if existing == nil || request == nil {
		return false
	}
	for _, requiredCheck := range request.Contexts {
		if !verifier.containsValue(existing.Contexts, requiredCheck) {
			return false
		}
	}
	return existing.Strict == request.Strict
}

func (verifier BranchProtectionVerifier) containsValue(values []string, value string) bool {
	for _, existingCheck := range values {
		if existingCheck == value {
			return true
		}
	}
	return false
}

func (verifier BranchProtectionVerifier) checkIfPrReviewPolicyIsApplied(existing *github.PullRequestReviewsEnforcement, request *github.PullRequestReviewsEnforcementRequest) bool {
	return existing != nil && request != nil &&
		existing.RequiredApprovingReviewCount >= request.RequiredApprovingReviewCount &&
		existing.DismissStaleReviews == request.DismissStaleReviews &&
		existing.RequireCodeOwnerReviews == request.RequireCodeOwnerReviews
}

func (verifier BranchProtectionVerifier) checkIfBranchRestrictionsAreApplied(existing *github.BranchRestrictions, request *github.BranchRestrictionsRequest) bool {
	return existing != nil && request != nil &&
		stringSlicesEqualIgnoringOrder(getTeamNames(existing.Teams), request.Teams) &&
		stringSlicesEqualIgnoringOrder(getUserNames(existing.Users), request.Users) &&
		stringSlicesEqualIgnoringOrder(getAppNames(existing.Apps), request.Apps)
}

func getTeamNames(teams []*github.Team) []string {
	var result []string
	for _, team := range teams {
		result = append(result, *team.Name)
	}
	return result
}

func getUserNames(users []*github.User) []string {
	var result []string
	for _, user := range users {
		result = append(result, *user.Name)
	}
	return result
}

func getAppNames(apps []*github.App) []string {
	var result []string
	for _, app := range apps {
		result = append(result, *app.Name)
	}
	return result
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

func (verifier BranchProtectionVerifier) createProtectionRequest(requireSonar bool) github.ProtectionRequest {
	allowForcePushes := false
	requiredChecks, err := verifier.getRequiredChecks(requireSonar)
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
		EnforceAdmins: true,
		Restrictions: &github.BranchRestrictionsRequest{
			Teams: []string{},
			Users: []string{},
			Apps:  []string{},
		},
		AllowForcePushes: &allowForcePushes,
	}
}

func (verifier BranchProtectionVerifier) getRequiredChecks(requireSonar bool) ([]string, error) {
	result := []string{}
	_, directory, _, err := verifier.client.Repositories.GetContents(context.Background(), "exasol", verifier.repoName, ".github/workflows/", &github.RepositoryContentGetOptions{})
	if err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "404 Not Found") {
			directory = []*github.RepositoryContent{}
		} else {
			return nil, err
		}
	}
	for _, fileDesc := range directory {
		workflowFilePath := fileDesc.Path
		if *fileDesc.Type == "dir" {
			continue
		}
		requiredChecksForWorkflow, err := verifier.getChecksForWorkflow(workflowFilePath)
		if err != nil {
			return nil, err
		}
		for _, check := range requiredChecksForWorkflow {
			if !verifier.containsValue(result, check) {
				result = append(result, check)
			}
		}
	}
	if requireSonar {
		result = append(result, "SonarCloud Code Analysis")
	}
	return result, nil
}

func (verifier BranchProtectionVerifier) getChecksForWorkflow(workflowFilePath *string) ([]string, error) {
	content, err := verifier.downloadFile(*workflowFilePath)
	if err != nil {
		return nil, err
	}
	return verifier.getChecksForWorkflowContent(content, workflowFilePath), nil
}

func (verifier BranchProtectionVerifier) getChecksForWorkflowContent(content string, fileName *string) []string {
	fileUrl := fmt.Sprintf("https://github.com/exasol/%s/blob/%s/%s", verifier.repoName, verifier.getRepo().GetDefaultBranch(), *fileName)
	workflow, err := WorkflowDefinitionParser{}.ParseWorkflowDefinition(content)
	if err != nil {
		switch err := err.(type) {
		case ValidationError:
			fmt.Printf("%vValidation Error for '%v': %v %v\n", consoleColorRed, fileUrl, err.Error(), consoleColorReset)
			os.Exit(1)
		default:
			fmt.Printf("%vWarning: Failed to parse workflow definition '%v'. Probably you use some advanced matrix build features there. Github-keeper will not add the checks from this workflow to the branch protection. Please add them manually. %v\n", consoleColorYellow, fileUrl, consoleColorReset)
			var emptyResult []string
			return emptyResult
		}
	}
	hasWorkflowPushOrPrTrigger := hasWorkflowPushOrPrTrigger(workflow.Trigger)
	if hasWorkflowPushOrPrTrigger {
		return workflow.JobsNames
	}
	var emptyResult []string
	return emptyResult
}

func hasWorkflowPushOrPrTrigger(triggers []string) bool {
	for _, trigger := range triggers {
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

type workflowDefinition struct {
	Name      string
	Trigger   []string
	JobsNames []string
}

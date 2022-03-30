package cmd

import (
	"context"
	"fmt"

	"github.com/google/go-github/v43/github"
)

type RepoSettingsVerifier struct {
	githubClient *github.Client
	repo         string
	org          string
}

type RepoProblemHandler interface {
	handleWrongSettings(template *github.Repository)
	handleMissingSecurityAlerts()
}

type FixRepoProblemHandler struct {
	githubClient *github.Client
	repo         string
	org          string
}

func (handler FixRepoProblemHandler) handleMissingSecurityAlerts() {
	handler.githubClient.Repositories.EnableVulnerabilityAlerts(context.Background(), handler.org, handler.repo)
}

func (handler FixRepoProblemHandler) handleWrongSettings(template *github.Repository) {
	_, _, err := handler.githubClient.Repositories.Edit(context.Background(), handler.org, handler.repo, template)
	if err != nil {
		panic(fmt.Sprintf("Failed to update repository settings for %v. Cause: %v", handler.repo, err.Error()))
	}
}

type LogRepoProblemHandler struct {
	repo string
}

func (handler LogRepoProblemHandler) handleMissingSecurityAlerts() {
	fmt.Printf("The repository %v does not enable Dependabot alerts.\n", handler.repo)
}

func (handler LogRepoProblemHandler) handleWrongSettings(template *github.Repository) {
	fmt.Printf("The repository %v has outdated repo settings.\n", handler.repo)
}

func (verifier *RepoSettingsVerifier) VerifyRepoSettings(fix bool) {
	problemHandler := verifier.getProblemHandler(fix)
	verifier.verifyBaseSetting(problemHandler)
	verifier.verifyVulnerabilityAlerts(problemHandler)
	verifier.enableDependabot(fix)
}

func (verifier *RepoSettingsVerifier) verifyBaseSetting(problemHandler RepoProblemHandler) {
	repo, _, err := verifier.githubClient.Repositories.Get(context.Background(), verifier.org, verifier.repo)
	if err != nil {
		panic(fmt.Sprintf("Failed to get settings for repository %v.\n", repo))
	}
	repositoryTemplate := verifier.getRepositoryTemplate()
	if verifier.checkIfRepoMatchesTemplate(repo, repositoryTemplate) {
		problemHandler.handleWrongSettings(&repositoryTemplate)
	}
}

func (verifier *RepoSettingsVerifier) enableDependabot(isFix bool) {
	/* Unfortunately the GitHub API does not support a GET for automated security fixes as of 2022-03. So we can't check
	if dependabot is enabled. For that reason, we decided to simply enable it in fix mode and don't validate.
	That's ok, since it's just a comfort feature. The security relevant feature are the alerts.	 */
	if isFix {
		_, err := verifier.githubClient.Repositories.EnableAutomatedSecurityFixes(context.Background(), verifier.org, verifier.repo)
		if err != nil {
			panic(fmt.Sprintf("Failed to enable security fixes for repository %v.\n.", verifier.repo))
		}
	}
}

func (verifier *RepoSettingsVerifier) verifyVulnerabilityAlerts(problemHandler RepoProblemHandler) {
	alertsEnabled, _, err := verifier.githubClient.Repositories.GetVulnerabilityAlerts(context.Background(), verifier.org, verifier.repo)
	if err != nil {
		panic(fmt.Sprintf("Failed to get securtiy alert status for repository %v.\n.", verifier.repo))
	}
	if !alertsEnabled {
		problemHandler.handleMissingSecurityAlerts()
	}
}

func (verifier *RepoSettingsVerifier) getProblemHandler(fix bool) RepoProblemHandler {
	if fix {
		return FixRepoProblemHandler{repo: verifier.repo, org: verifier.org, githubClient: verifier.githubClient}
	} else {
		return LogRepoProblemHandler{repo: verifier.repo}
	}
}

func (verifier *RepoSettingsVerifier) checkIfRepoMatchesTemplate(repo *github.Repository, repositoryTemplate github.Repository) bool {
	return *repo.AllowAutoMerge != *repositoryTemplate.AllowAutoMerge || *repo.DeleteBranchOnMerge != *repositoryTemplate.DeleteBranchOnMerge
}

func (verifier *RepoSettingsVerifier) getRepositoryTemplate() github.Repository {
	allowAutoMerge := true
	deleteBranchOnMerge := true
	repositoryRequest := github.Repository{AllowAutoMerge: &allowAutoMerge, DeleteBranchOnMerge: &deleteBranchOnMerge}
	return repositoryRequest
}

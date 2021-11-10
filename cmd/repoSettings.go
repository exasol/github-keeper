package cmd

import (
	"context"
	"fmt"
	"github.com/google/go-github/v39/github"
)

type RepoSettingsVerifier struct {
	githubClient *github.Client
	repo         string
	org          string
}

type RepoProblemHandler interface {
	handleWrongSettings(template *github.Repository)
}

type FixRepoProblemHandler struct {
	githubClient *github.Client
	repo         string
	org          string
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

func (handler LogRepoProblemHandler) handleWrongSettings(template *github.Repository) {
	fmt.Printf("The repository %v has outdated repo settings.\n", handler.repo)
}

func (verifier *RepoSettingsVerifier) VerifyRepoSettings(fix bool) {
	problemHandler := verifier.getProblemHandler(fix)

	repo, _, err := verifier.githubClient.Repositories.Get(context.Background(), verifier.org, verifier.repo)
	if err != nil {
		panic(fmt.Sprintf("Failed to get settings for repository %v", repo))
	}
	repositoryTemplate := verifier.getRepositoryTemplate()
	if verifier.checkIfRepoMatchesTemplate(repo, repositoryTemplate) {
		problemHandler.handleWrongSettings(&repositoryTemplate)
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

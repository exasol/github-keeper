package cmd

import (
	"context"
	"fmt"
	"github.com/google/go-github/v39/github"
)

type WebHookProblemHandler interface {
	createHook(template *github.Hook)
	updateHook(existing *github.Hook, template *github.Hook)
}

type LogWebHookProblemHandler struct {
	repo string
}

func (handler LogWebHookProblemHandler) createHook(template *github.Hook) {
	fmt.Printf("Missing required web hook '%v' for repository %v.\n", *template.Name, handler.repo)
}

func (handler LogWebHookProblemHandler) updateHook(existing *github.Hook, template *github.Hook) {
	fmt.Printf("Outdated web hook '%v' for repository %v.\n", *template.Name, handler.repo)
}

type FixWebHookProblemHandler struct {
	githubClient *github.Client
	repo         string
	org          string
}

func (handler FixWebHookProblemHandler) createHook(template *github.Hook) {
	_, _, err := handler.githubClient.Repositories.CreateHook(context.Background(), handler.org, handler.repo, template)
	if err != nil {
		panic(fmt.Sprintf("Failed to create web-hook for repo %v. Cause: %v", handler.repo, err.Error()))
	}
}

func (handler FixWebHookProblemHandler) updateHook(existing *github.Hook, template *github.Hook) {
	_, _, err := handler.githubClient.Repositories.EditHook(context.Background(), handler.org, handler.repo, existing.GetID(), template)
	if err != nil {
		panic(fmt.Sprintf("Failed to update web-hook %v for repo %v. Cause: %v", existing.Name, handler.repo, err.Error()))
	}
}

func (verifier *WebHookVerifier) VerifyWebHooks(fix bool) {
	problemHandler := verifier.getProblemHandler(fix)
	hooks, _, err := verifier.githubClient.Repositories.ListHooks(context.Background(), verifier.org, verifier.repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		panic(fmt.Sprintf("Failed to list web-hooks for repository %v. Cause: %v", verifier.repo, err.Error()))
	}
	hookTemplate := verifier.createIssuesHookTemplate()
	url := hookTemplate.Config["url"].(string)
	hook := verifier.findHookByUrl(hooks, &url)
	if hook == nil {
		problemHandler.createHook(hookTemplate)
	} else {
		if !verifier.checkIfHookMatchesTemplate(hook, hookTemplate) {
			problemHandler.updateHook(hook, hookTemplate)
		}
	}
}

func (verifier *WebHookVerifier) getProblemHandler(fix bool) WebHookProblemHandler {
	if fix {
		return FixWebHookProblemHandler{
			org:          verifier.org,
			repo:         verifier.repo,
			githubClient: verifier.githubClient,
		}
	} else {
		return LogWebHookProblemHandler{verifier.repo}
	}
}

func (verifier *WebHookVerifier) findHookByUrl(hooks []*github.Hook, url *string) *github.Hook {
	for _, hook := range hooks {
		if hook.Config["url"].(string) == *url {
			return hook
		}
	}
	return nil
}

func (verifier *WebHookVerifier) checkIfHookMatchesTemplate(hook *github.Hook, issuesHook *github.Hook) bool {
	return *hook.Active == *issuesHook.Active &&
		hook.Config["url"] == issuesHook.Config["url"] &&
		hook.Config["content_type"] == issuesHook.Config["content_type"] &&
		stringSlicesEqualIgnoringOrder(hook.Events, issuesHook.Events)
}

func (verifier *WebHookVerifier) createIssuesHookTemplate() *github.Hook {
	active := true
	name := "Issues on Slack"
	url := verifier.secrets.resolveSecret("issuesSlackWebhookUrl")
	hook := github.Hook{
		Events: []string{"release", "issues", "repository_vulnerability_alert", "secret_scanning_alert", "repository"},
		Active: &active,
		Name:   &name,
		Config: map[string]interface{}{
			"content_type": "form",
			"url":          url,
		},
	}
	return &hook
}

type WebHookVerifier struct {
	githubClient *github.Client
	repo         string
	org          string
	secrets      *Secrets
}

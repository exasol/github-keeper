package cmd

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-github/v43/github"
	"github.com/stretchr/testify/suite"
)

type WebHooksSuite struct {
	IntegrationTestSuite
	verifier       *WebHookVerifier
	testWebhookUrl string
}

func TestWebHooksSuite(t *testing.T) {
	suite.Run(t, new(WebHooksSuite))
}

func (suite *WebHooksSuite) SetupSuite() {
	suite.IntegrationTestSuite.SetupSuite()
	suite.testWebhookUrl = "https://slack.com/123"
	suite.verifier = &WebHookVerifier{githubClient: suite.githubClient, repo: suite.testRepo, org: suite.testOrg,
		secrets: &Secrets{secrets: map[string]string{"issuesSlackWebhookUrl": suite.testWebhookUrl}}}
	suite.deleteAllHooks()
}

func (suite WebHooksSuite) TearDownTest() {
	suite.deleteAllHooks()
}

func (suite WebHooksSuite) TestMissingWebhookMessage() {
	output := suite.CaptureOutput(func() {
		suite.verifier.VerifyWebHooks(false)
	})
	suite.Assert().Equal(output, "Missing required web hook 'Issues on Slack' for repository testing-release-robot.\n")
}

func (suite WebHooksSuite) TestCreateWebhook() {
	suite.verifier.VerifyWebHooks(true)
	suite.assertWebhook()
}

func (suite WebHooksSuite) TestIncorrectWebhook() {
	suite.createIncorrectWebhook()
	output := suite.CaptureOutput(func() {
		suite.verifier.VerifyWebHooks(false)
	})
	suite.Assert().Equal("Outdated web hook 'Issues on Slack' for repository testing-release-robot.\n", output)
}

func (suite WebHooksSuite) TestUpdateWebhook() {
	suite.createIncorrectWebhook()
	suite.verifier.VerifyWebHooks(true)
	suite.assertWebhook()
}

func (suite WebHooksSuite) createIncorrectWebhook() {
	active := false
	name := "Issues on Slack"
	hook := github.Hook{
		Events: []string{"release"},
		Active: &active,
		Name:   &name,
		Config: map[string]interface{}{
			"content_type": "json",
			"url":          suite.testWebhookUrl,
		},
	}
	_, _, err := suite.githubClient.Repositories.CreateHook(context.Background(), suite.testOrg, suite.testRepo, &hook)
	suite.NoError(err)
}

func (suite WebHooksSuite) assertWebhook() {
	hooks, _, err := suite.githubClient.Repositories.ListHooks(context.Background(), suite.testOrg, suite.testRepo, &github.ListOptions{PerPage: 100})
	suite.NoError(err)
	hook := suite.findTestWebhook(hooks)
	suite.NotNil(hook)
	expected := []string{"release", "issues", "repository_vulnerability_alert", "secret_scanning_alert", "repository"}
	sort.Strings(hook.Events)
	sort.Strings(expected)
	suite.Equal(expected, hook.Events)
	suite.Assert().True(*hook.Active)
	suite.Assert().Equal("form", hook.Config["content_type"])
}

func (suite WebHooksSuite) findTestWebhook(hooks []*github.Hook) *github.Hook {
	for _, hook := range hooks {
		if hook.Config["url"].(string) == suite.testWebhookUrl {
			return hook
		}
	}
	return nil
}

func (suite WebHooksSuite) deleteAllHooks() {
	hooks, _, err := suite.githubClient.Repositories.ListHooks(context.Background(), suite.testOrg, suite.testRepo, &github.ListOptions{PerPage: 100})
	suite.NoError(err)
	for _, hook := range hooks {
		_, err := suite.githubClient.Repositories.DeleteHook(context.Background(), suite.testOrg, suite.testRepo, hook.GetID())
		suite.NoError(err)
	}
}

package cmd

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type SecretsSuite struct {
	suite.Suite
}

func TestSecretsSuite(t *testing.T) {
	suite.Run(t, new(SecretsSuite))
}

func (suite SecretsSuite) TestRead() {
	secrets := ReadSecretsFromYaml("../test_resources/secrets.yml")
	suite.Assert().Equal("https://slack.com/123", secrets.resolveSecret("issuesSlackWebhookUrl"))
}

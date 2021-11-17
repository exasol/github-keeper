package cmd

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ConfigureRepoSuite struct {
	IntegrationTestSuite
}

func TestConfigureRepoSuite(t *testing.T) {
	suite.Run(t, new(ConfigureRepoSuite))
}

func (suite *ConfigureRepoSuite) SetupSuite() {
	suite.IntegrationTestSuite.SetupSuite()
}

func (suite *ConfigureRepoSuite) TestRunCommand() {
	output := suite.CaptureOutput(func() {
		err := configureRepoCmd.Flags().Set("secrets", "../test_resources/secrets.yml")
		suite.NoError(err)
		configureRepoCmd.Run(configureRepoCmd, []string{suite.testRepo})
	})
	suite.Assert().Contains(output, suite.testRepo)
}

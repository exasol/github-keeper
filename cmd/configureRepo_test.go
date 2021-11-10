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
	output := suite.captureOutput(func() {
		configureRepoCmd.Run(configureRepoCmd, []string{suite.testRepo})
	})
	suite.Assert().Contains(output, suite.testRepo)
}

package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type WorkflowDefinitionParserSuite struct {
	suite.Suite
}

func TestWorkflowDefinitionParserSuite(t *testing.T) {
	suite.Run(t, new(WorkflowDefinitionParserSuite))
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithListSyntax() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  - push
jobs:
  build:
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 1)
	suite.Contains(jobNames, "build")
	suite.Equal(true, definition.Trigger.TriggerOnPushToAnyBranch)
	suite.Contains(definition.Name, "CI Build")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMapSyntaxAndSpecificBranch() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
    branches:
      - main
jobs:
  build:
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	suite.Equal(false, definition.Trigger.TriggerOnPushToAnyBranch)
	suite.Equal(false, definition.Trigger.TriggerOnPr)
	suite.Contains(definition.Trigger.TriggerOnPushToBranches, "main")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithPrTriggerAndMapSyntax() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  pull_request:
jobs:
  build:
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	suite.Equal(false, definition.Trigger.TriggerOnPushToAnyBranch)
	suite.Equal(true, definition.Trigger.TriggerOnPr)
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMapSyntax() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 1)
	suite.Contains(jobNames, "build")
	suite.Equal(true, definition.Trigger.TriggerOnPushToAnyBranch)
	suite.Contains(definition.Name, "CI Build")
}

func (suite *WorkflowDefinitionParserSuite) TestParseTriggerWithSchedule() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  schedule:
    - cron: "0 5 * * *"
  push:
jobs:
  build:
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	suite.Equal(true, definition.Trigger.TriggerOnPushToAnyBranch)
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithJobName() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    name: My-Job
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 1)
	suite.Contains(jobNames, "My-Job")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMatrixBuild() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        a: [ "1", "2"]
        b: [ "3", "4" ]
    name: Build with A ${{ matrix.a }} and B ${{ matrix.b }}
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 4)
	suite.Contains(jobNames, "Build with A 1 and B 3")
	suite.Contains(jobNames, "Build with A 1 and B 4")
	suite.Contains(jobNames, "Build with A 2 and B 3")
	suite.Contains(jobNames, "Build with A 2 and B 4")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMatrixBuildWithMultiplesParameters() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        a:
         - id: 1
           num: 10
         - id: 2
           num: 20
        b: [ "3" ]
    name: Build with id ${{ matrix.id }}, num ${{matrix.num}} and B ${{ matrix.b }}
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 2)
	suite.Contains(jobNames, "Build with id 1, num 10 and B 3")
	suite.Contains(jobNames, "Build with id 2, num 20 and B 3")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMatrixBuildWithConfigSyntax() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        config:
          - {a: 10, b: 3}
          - {a: 20, b: 3}
    name: Build with A ${{ matrix.a }} and B ${{ matrix.b }}
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 2)
	suite.Contains(jobNames, "Build with A 10 and B 3")
	suite.Contains(jobNames, "Build with A 20 and B 3")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMultiDimensionMatrixBuildAndNoName() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        a: [1, 2]
        b: [ "3" ]
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	_, err = definition.GetJobNames()
	suite.Equal("multi dimensional matrix github-action jobs with no explicit name are not supported. Please add a name field to the job that combines the matrix parameters into a more readable name. For example \"Build with Go ${{matrix.go}} and Exasol ${{ matrix.db }}\"", err.Error())
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMatrixBuildWithMultiplesParametersAndNoName() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        a:
         - id: 1
           num: 10
         - id: 2
           num: 20
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	_, err = definition.GetJobNames()
	suite.Equal("matrix github-action jobs with object parameters and no job name are not supported. Please add a name field to the job that combines the matrix parameters into a more readable name. For example \"Build with Go ${{matrix.go}} and Exasol ${{ matrix.db }}\"", err.Error())
}

func (suite *WorkflowDefinitionParserSuite) TestMatrixBuildWithoutNameIgnoredForNonPullRequestWorkflows() {
	parser := WorkflowDefinitionParser{}
	_, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  other:
jobs:
  build:
    strategy:
      matrix:
        a:
         - id: 1
           num: 10
         - id: 2
           num: 20
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMatrixBuildAndNoName() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        a: [1,2]
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 2)
	suite.Contains(jobNames, "build (1)")
	suite.Contains(jobNames, "build (2)")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMatrixBuildAndFloatValue() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        a: [1.2,2.1]
    runs-on: ubuntu-latest
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 2)
	suite.Contains(jobNames, "build (1.2)")
	suite.Contains(jobNames, "build (2.1)")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithMatrixNotAllVariablesUsed() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
on:
  push:
  pull_request:
  workflow_dispatch:

jobs:
  examples:
    strategy:
      fail-fast: false
      matrix:
        include:
          - python-version: 3.9
            exasol-tag: latest-7.1
            exasol-port: 8563

          - python-version: 3.95
            exasol-tag: "latest-7.0"
            exasol-port: 8563

          - python-version: "3.6"
            exasol-tag: latest-6.2
            exasol-port: 8888
    name: "Build with Python ${{ matrix.python-version }} and Exasol ${{ matrix.exasol-tag }}"
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 3)
	suite.Contains(jobNames, "Build with Python 3.9 and Exasol latest-7.1")
	suite.Contains(jobNames, "Build with Python 3.95 and Exasol latest-7.0")
	suite.Contains(jobNames, "Build with Python 3.6 and Exasol latest-6.2")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowContentWithSingleItemMatrix() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
on:
  push:
  pull_request:
  workflow_dispatch:

jobs:
  ssl_cert:
    strategy:
      fail-fast: false
      matrix:
        include:
          - python-version: 3.9
            exasol-tag: 7.1.6
    name: "Build SSL-Cert with Python ${{ matrix.python-version }} and Exasol ${{ matrix.exasol-tag }}"
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 1)
	suite.Contains(jobNames, "Build SSL-Cert with Python 3.9 and Exasol 7.1.6")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowForUnsupportedSyntax() {
	parser := WorkflowDefinitionParser{}
	_, err := parser.ParseWorkflowDefinition(`
name: CI Build
on:
  push:
jobs:
  build:
    strategy:
      matrix:
        test-path: ${{fromJson(needs.prep-testbed.outputs.matrix)}}
    runs-on: ubuntu-latest
`)
	suite.Assert().Error(err)
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowWithFloatParameter() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
  name: CI-CD
  on:
    push:
  jobs:
    run_tests:
      runs-on: ubuntu-latest
      strategy:
        fail-fast: false
        matrix:
          python: [3.6, 3.72, 3.86]
          exasol_version:
            - 7.1.6
      name: Run Tests (Python-${{ matrix.python }}, Exasol-${{ matrix.exasol_version }})
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 3)
	suite.Contains(jobNames, "Run Tests (Python-3.6, Exasol-7.1.6)")
	suite.Contains(jobNames, "Run Tests (Python-3.7, Exasol-7.1.6)")
	suite.Contains(jobNames, "Run Tests (Python-3.9, Exasol-7.1.6)")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowWithIntParameter() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
  name: CI-CD
  on:
    push:
  jobs:
    run_tests:
      runs-on: ubuntu-latest
      strategy:
        fail-fast: false
        matrix:
          python: [3]
          exasol_version:
            - 7.1.6
      name: Run Tests (Python-${{ matrix.python }}, Exasol-${{ matrix.exasol_version }})
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 1)
	suite.Contains(jobNames, "Run Tests (Python-3, Exasol-7.1.6)")
}

func (suite *WorkflowDefinitionParserSuite) TestGetChecksForWorkflowWithBoolParameter() {
	parser := WorkflowDefinitionParser{}
	definition, err := parser.ParseWorkflowDefinition(`
  name: CI-CD
  on:
    push:
  jobs:
    run_tests:
      runs-on: ubuntu-latest
      strategy:
        fail-fast: false
        matrix:
          success: [true]
          exasol_version:
            - 7.1.6
      name: Run Tests (Success-${{ matrix.success }}, Exasol-${{ matrix.exasol_version }})
`)
	suite.NoError(err)
	jobNames, err := definition.GetJobNames()
	suite.NoError(err)
	suite.Len(jobNames, 1)
	suite.Contains(jobNames, "Run Tests (Success-true, Exasol-7.1.6)")
}

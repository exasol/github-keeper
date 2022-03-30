package cmd

import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"

	"github.com/google/go-github/v43/github"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	githubClient      *github.Client
	testOrg           string
	testRepo          string
	testDefaultBranch string
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.testOrg = "exasol"
	suite.testRepo = "testing-release-robot"
	suite.testDefaultBranch = "master"
	suite.githubClient = getGithubClient()
}

func (suite *IntegrationTestSuite) CaptureOutput(functionToCapture func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	isReaderReady := new(sync.WaitGroup)
	isReaderReady.Add(1)
	go func() {
		var buffer bytes.Buffer
		isReaderReady.Done()
		_, err := io.Copy(&buffer, reader) //blocking
		if err != nil {
			panic(err)
		}
		out <- buffer.String()
	}()
	isReaderReady.Wait()
	functionToCapture()
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	return <-out
}

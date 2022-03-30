package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/alyu/configparser"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

func getGithubClient() *github.Client {
	githubToken := readGithubTokenFromConfig()
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func readGithubTokenFromConfig() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user's home directory. Cause: %v", err.Error()))
	}
	configFile := path.Join(homedir, ".release-droid", "credentials")
	configparser.Delimiter = "="
	config, err := configparser.Read(configFile)
	if err != nil {
		panic(err)
	}
	oauthToken, err := config.StringValue("global", "github_oauth_access_token")
	if err != nil {
		panic(fmt.Sprintf("The config file (%v) did not contain the required 'github_oauth_access_token' value.", configFile))
	}
	return oauthToken
}

package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var configureRepoCmd = &cobra.Command{
	Use:   "configure-repo <repo-name> [more repo names]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Verify the config of a given repository",
	Run: func(cmd *cobra.Command, args []string) {
		client := getGithubClient()
		fix, err := cmd.Flags().GetBool("fix")
		if err != nil {
			panic(fmt.Sprintf("Could not read parameter fix: %v", err.Error()))
		}
		secretsFile, err := cmd.Flags().GetString("secrets")
		if err != nil {
			panic(fmt.Sprintf("Could not read parameter secrets: %v", err.Error()))
		}
		secrets := ReadSecretsFromYaml(secretsFile)
		for _, repo := range args {
			fmt.Printf("\nhttps://github.com/exasol/%v\n", repo)
			branchProtectionVerifier := BranchProtectionVerifier{client: client, repoName: repo}
			branchProtectionVerifier.CheckIfBranchProtectionIsApplied(fix)
			UnifyLabels(repo, client, fix)
			settingsVerifier := RepoSettingsVerifier{repo: repo, githubClient: client, org: "exasol"}
			settingsVerifier.VerifyRepoSettings(fix)
			webHookVerifier := WebHookVerifier{secrets: secrets, repo: repo, githubClient: client, org: "exasol"}
			webHookVerifier.VerifyWebHooks(fix)
		}
	},
}

func getDefaultConfigFile() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user's home directory. Cause: %v", err.Error()))
	}
	return path.Join(homedir, ".github-keeper", "secrets.yml")
}

func init() {
	configureRepoCmd.Flags().Bool("fix", false, "If this flag is set, github-keeper fixed the findings. Otherwise it just prints the diff.")
	configureRepoCmd.Flags().String("secrets", getDefaultConfigFile(), "Use a different secrets file location")
	rootCmd.AddCommand(configureRepoCmd)
}

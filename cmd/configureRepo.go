package cmd

import (
	"fmt"
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
		for _, repo := range args {
			fmt.Printf("\n%v \n", repo)
			verifier := BranchProtectionVerifier{client: client, repoName: repo}
			verifier.CheckIfBranchProtectionIsApplied(fix)
			UnifyLabels(repo, client, fix)
		}
	},
}

func init() {
	configureRepoCmd.Flags().Bool("fix", false, "If this flag is set, github-keeper fixed the findings. Otherwise it just prints the diff.")
	rootCmd.AddCommand(configureRepoCmd)
}

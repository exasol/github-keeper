package cmd

import (
	"context"
	"fmt"

	"github.com/google/go-github/v39/github"
	"github.com/spf13/cobra"
)

var reactivateScheduledActionsCmd = &cobra.Command{
	Use:   "reactivate-scheduled-github-actions <repo-name>",
	Args:  cobra.MinimumNArgs(1),
	Short: "Reactivate the scheduled GitHub actions for the given repository.",
	Long:  "GitHub automatically disables the run of scheduled actions after some time. This tool helps you to reenable them.",
	Run: func(cmd *cobra.Command, args []string) {
		client := getGithubClient()
		for _, repo := range args {
			reEnableWorkflows(repo, client)
		}
	},
}

func reEnableWorkflows(repoName string, client *github.Client) {
	workflows, _, err := client.Actions.ListWorkflows(context.Background(), "exasol", repoName, &github.ListOptions{PerPage: 1000})
	if err != nil {
		panic(fmt.Sprintf("Failed to list the workflows of %s. Cause: %s", repoName, err.Error()))
	}
	for _, workflow := range workflows.Workflows {
		if *workflow.State != "active" {
			fmt.Printf("Reactivating %v/%v\n", repoName, *workflow.Name)
			_, err := client.Actions.EnableWorkflowByID(context.Background(), "exasol", repoName, *workflow.ID)
			if err != nil {
				panic(fmt.Sprintf("Failed to re-enable workflow '%s' of repository '%s'. Cause: %s", *workflow.Name, repoName, err.Error()))
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(reactivateScheduledActionsCmd)
}

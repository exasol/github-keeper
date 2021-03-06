package cmd

import (
	"context"
	"fmt"

	"github.com/google/go-github/v43/github"
	"github.com/spf13/cobra"
)

var listMyReposCmd = &cobra.Command{
	Use:   "list-my-repos",
	Short: "List all repositories of the Exasol organization where I'm the admin and that are not archived.",
	Run: func(cmd *cobra.Command, args []string) {
		client := getGithubClient()
		opt := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}
		for {
			repos, resp, err := client.Repositories.ListByOrg(context.Background(), "exasol", opt)
			if err != nil {
				panic("Failed to list repositories. Cause: " + err.Error())
			}
			for _, repo := range repos {
				if (repo.Permissions)["admin"] && !*repo.Archived {
					fmt.Print(" " + *repo.Name)
				}
			}
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	},
}

func init() {
	rootCmd.AddCommand(listMyReposCmd)
}

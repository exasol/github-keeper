package cmd

import (
	"context"
	"fmt"

	"github.com/google/go-github/v39/github"
	"github.com/spf13/cobra"
)

func getLabelModifier(fix bool, repo string, githubClient *github.Client) LablesModifier {
	if fix {
		return &RealLabelModifier{repo: repo, githubClient: githubClient}
	} else {
		return &DryRunLabelModifier{}
	}
}

func unifyLabels(repo string, githubClient *github.Client, fix bool) {
	println("\n" + repo)
	labelDefinitions := []*LabelDesc{
		{"feature", "88ee66", []string{"enhancement"}, true},
		{"bug", "ee0000", []string{}, true},
		{"documentation", "0000ee", []string{}, true},
		{"refactoring", "ffbb11", []string{}, true},
		{"duplicate", "cccccc", []string{}, true},
		{"invalid", "eeeeee", []string{}, true},
		{"question", "cc3377", []string{"help wanted"}, true},
		{"ci", "cc3377", []string{}, false},
		{"source:exasol", "eeeeee", []string{}, true},
		{"source:external", "eeeeee", []string{}, true},
		{"decision:wont-fix", "ffffff", []string{"wontfix", "won't fix", "status:wont-fix"}, true},
		{"shelved:yes", "ff33cc", []string{}, true},
		{"timeline:long-term", "555555", []string{"long-term", "timeline:longterm", "timelien:long-term"}, true},
		{"complexity:low", "4FC24F", []string{"good-first-issue"}, true},
		{"complexity:medium", "F2BF63", []string{}, true},
		{"complexity:high", "F26363", []string{}, true},
		{"dependencies", "ffbb11", []string{}, false},
		{"security", "ee0000", []string{}, false}, //check if we can configure
		{"blocked:yes", "000000", []string{"blocked", "status:blocked"}, true}}
	labelModifier := getLabelModifier(fix, repo, githubClient)
	labels := listLabels(repo, githubClient)
	for _, label := range labels {
		labelDesc := findLabelDefinitionByName(*label.Name, labelDefinitions)
		if labelDesc == nil {
			labelDescByOldName := findLabelDefinitionByOldName(*label.Name, labelDefinitions)
			if labelDescByOldName == nil {
				labelModifier.removeLabel(label)
			} else {
				labelModifier.renameLabel(label, labelDescByOldName)
			}
		}
	}
	labels = listLabels(repo, githubClient) // list again to get renamed
	for _, labelDefinition := range labelDefinitions {
		label := findLabelByName(labelDefinition.name, labels)
		if label == nil {
			if labelDefinition.required {
				labelModifier.createLabel(labelDefinition)
			}
		} else {
			if *label.Color != labelDefinition.color {
				labelModifier.setColor(label, labelDefinition)
			}
		}
	}
}

func listLabels(repo string, githubClient *github.Client) []*github.Label {
	labels, _, err := githubClient.Issues.ListLabels(context.Background(), "exasol", repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		panic("Failed to list labels")
	}
	return labels
}

type LablesModifier interface {
	createLabel(labelDefinition *LabelDesc)
	removeLabel(label *github.Label)
	renameLabel(label *github.Label, labelDefinition *LabelDesc)
	setColor(label *github.Label, labelDefinition *LabelDesc)
}

type DryRunLabelModifier struct {
}

func (dryRunModifier *DryRunLabelModifier) createLabel(labelDefinition *LabelDesc) {
	fmt.Printf("Missing required label '%s'. Would create.\n", labelDefinition.name)
}

func (dryRunModifier *DryRunLabelModifier) removeLabel(label *github.Label) {
	fmt.Printf("Superfluous label '%s'. Would remove.\n", *label.Name)
}

func (dryRunModifier *DryRunLabelModifier) renameLabel(label *github.Label, labelDefinition *LabelDesc) {
	fmt.Printf("The label '%s' was renamed to '%s'. Would rename.\n", *label.Name, labelDefinition.name)
}

func (dryRunModifier *DryRunLabelModifier) setColor(label *github.Label, labelDefinition *LabelDesc) {
	fmt.Printf("Label '%s' has wrong color %s. Expected: %s. Would change.\n", *label.Name, *label.Color, labelDefinition.color)
}

type RealLabelModifier struct {
	githubClient *github.Client
	repo         string
}

func (realRunModifer *RealLabelModifier) createLabel(labelDefinition *LabelDesc) {
	_, _, err := realRunModifer.githubClient.Issues.CreateLabel(context.Background(), "exasol", realRunModifer.repo, &github.Label{Name: &labelDefinition.name, Color: &labelDefinition.color})
	if err != nil {
		panic(fmt.Sprintf("Failed to create label '%s' for repo '%s'. Cause: '%s'", labelDefinition.name, realRunModifer.repo, err.Error()))
	}
}

func (realRunModifer *RealLabelModifier) removeLabel(label *github.Label) {
	_, err := realRunModifer.githubClient.Issues.DeleteLabel(context.Background(), "exasol", realRunModifer.repo, *label.Name)
	if err != nil {
		panic(fmt.Sprintf("Failed to delete label '%s' for repo '%s'. Cause: '%s'", *label.Name, realRunModifer.repo, err.Error()))
	}
}

func (realRunModifer *RealLabelModifier) renameLabel(label *github.Label, labelDefinition *LabelDesc) {
	err := realRunModifer.updateLabel(label, labelDefinition)
	if err != nil {
		panic(fmt.Sprintf("Failed to rename label '%s' for repo '%s'. Cause: '%s'", *label.Name, realRunModifer.repo, err.Error()))
	}
}

func (realRunModifer *RealLabelModifier) setColor(label *github.Label, labelDefinition *LabelDesc) {
	err := realRunModifer.updateLabel(label, labelDefinition)
	if err != nil {
		panic(fmt.Sprintf("Failed to change color of label '%s' for repo '%s'. Cause: '%s'", *label.Name, realRunModifer.repo, err.Error()))
	}
}

func (realRunModifer *RealLabelModifier) updateLabel(label *github.Label, labelDefinition *LabelDesc) error {
	oldName := *label.Name
	label.Name = &labelDefinition.name
	label.Color = &labelDefinition.color
	_, _, err := realRunModifer.githubClient.Issues.EditLabel(context.Background(), "exasol", realRunModifer.repo, oldName, label)
	return err
}

func findLabelByName(name string, labels []*github.Label) *github.Label {
	for _, label := range labels {
		if *label.Name == name {
			return label
		}
	}
	return nil
}

func findLabelDefinitionByName(name string, labelDefinitions []*LabelDesc) *LabelDesc {
	for _, labelDescription := range labelDefinitions {
		if labelDescription.name == name {
			return labelDescription
		}
	}
	return nil
}

func findLabelDefinitionByOldName(name string, labelDefinitions []*LabelDesc) *LabelDesc {
	for _, labelDescription := range labelDefinitions {
		for _, oldName := range labelDescription.oldNames {
			if oldName == name {
				return labelDescription
			}
		}
	}
	return nil
}

// unifyLabelsCmd represents the unifyLabels command
var unifyLabelsCmd = &cobra.Command{
	Use:   "unify-labels <repo name>",
	Short: "Unifies the GitHub labels for a given repo",
	Args:  cobra.MinimumNArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		githubClient := getGithubClient()
		fix, err := cmd.Flags().GetBool("fix")
		if err != nil {
			panic("Could not read parameter fix")
		}
		for _, repo := range args {
			unifyLabels(repo, githubClient, fix)
		}
	},
}

func init() {
	unifyLabelsCmd.Flags().Bool("fix", false, "If this flag is set github-keeper unifies the labels. Otherwise it just prints the diff.")
	rootCmd.AddCommand(unifyLabelsCmd)
}

type LabelDesc struct {
	name     string
	color    string
	oldNames []string
	required bool
}

# GitHub Keeper

GitHub keeper is a CLI tool that helps to unify our repositories.

Features:

* List all Exasol repos where you are admin (`list-my-repos`)
* Create / Update the branch protection for the default branch according to our standards (`create-branch-protection`)
* Unify the labels of a GitHub repository (`unify-labels`)

## Installation

1. Install Go language. Minimum version is 1.13.
    * On Debian / Ubuntu:
        ```sh
        sudo apt install golang-go
        ```
2. Install dependendent packages: 
    ```
    cd $HOME/<PATH TO THIS REPO>/
    go get ./...
    ```

## Usage

in the github-keeper directory run:

```shell
go run .
```

Hint: For setting up a branch protection for all your repos use:

```shell
go run . create-branch-protection $(go run . list-my-repos)
```

### Tips

Run all repos of the integration team:

```shell
gh repo list exasol --limit 500 --json name,repositoryTopics --jq '.[] | select(.repositoryTopics.[]?.name == "exasol-integration") | .name' | tr "\n" " "
```

Check if a repo has disabled GitHub Actions:

```shell
gh repo list exasol --limit 500 --json name,repositoryTopics --jq '.[] | select(.repositoryTopics.[]?.name == "exasol-integration") | .name' | tr "\n" " " | xargs go run . reactivate-scheduled-github-actions
```

## Additional Information

* [Changelog](doc/changes/changelog.md)

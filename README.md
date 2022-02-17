# GitHub Keeper

GitHub keeper is a CLI tool that helps to unify our repositories.

[![CI Build](https://github.com/exasol/github-keeper/actions/workflows/ci-build.yml/badge.svg)](https://github.com/exasol/github-keeper/actions/workflows/ci-build.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=com.exasol%3Agithub-keeper&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=com.exasol%3Agithub-keeper)

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

## Configuration

Please create the config file `~/.github-keeper/secrets.yml` with the following content:

``` yaml
issuesSlackWebhookUrl: "<SLACK WEB-HOOK URL FOR ISSUE UPDATES>"
```

Hint: Instead of collecting all variables by hand you can also copy the file-content from the integration-teams keeper
vault.

## Usage

in the github-keeper directory run:

```shell
go run .
```

Hint: To verify the setup of all your repos use:

```shell
go run . configure-repo $(go run . list-my-repos)
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

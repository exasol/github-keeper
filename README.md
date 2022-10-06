# GitHub Keeper

GitHub keeper is a CLI tool that helps to unify our repositories.

[![CI Build](https://github.com/exasol/github-keeper/actions/workflows/ci-build.yml/badge.svg)](https://github.com/exasol/github-keeper/actions/workflows/ci-build.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=com.exasol%3Agithub-keeper&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=com.exasol%3Agithub-keeper)

## Features

* Reactivate scheduled GitHub actions that have been automatically disabled by GitHub after some time.
* List all Exasol repositories that are not archived and where you are admin.
* Inspect and fix settings of a GitHub repository regarding labels, branch protection and actions.

## Installation

1. Install Go language. Minimum version is 1.13.
   * On Debian / Ubuntu:
       ```shell
       sudo apt install golang-go
       ```
   * On macOS:
       ```shell
       brew install golang
       ```
2. Install dependendent packages:
    ```shell
    cd $HOME/<PATH TO THIS REPO>/
    go get ./...
    ```

## Configuration

Please create the configuration file `~/.github-keeper/secrets.yml` with the following content:

```yaml
issuesSlackWebhookUrl: "<slack web-hook url for issue updates>"
```

The url can be found in the keeper-vault.

## Usage

In the github-keeper directory run:

```shell
go run .
```

| Command | Description |
|---------|-------------|
| `go run . help command` | Display command line help |
| `go run .  reactivate-scheduled-github-actions`|  Reactivate the scheduled GitHub actions for the given repository. |
| `go run . completion <shell>` | Generate autocompletion script for shell `<shell>` |
| `go run . list-my-repos` | List all Exasol repositories where you are admin |
| `go run . configure-repo <repo-name> [more repo names] [flags]` | Inspect settings of GitHub repository |

### `list-my-repos`
List all repositories of the Exasol organization where I'm the admin and that are not archived.

Usage: `github-keeper list-my-repos [flags]`

| Flags | Description |
|-------|-------------|
|  `-h`, `--help`  | help |

### `reactivate-scheduled-github-actions`

Reenable scheduled actions automatically disabled by GitHub after some time.

Usage: `github-keeper reactivate-scheduled-github-actions <repo-name> [flags]`

| Flags | Description |
|-------|-------------|
|  `-h`, `--help`  | help |

### `configure-repo`
Verify the config of a given repository

Usage: `github-keeper configure-repo <repo-name> [more repo names] [flags]`

| Flags | Description |
|-------|-------------|
| `--fix` | If this flag is set, github-keeper fixed the findings. Otherwise it just prints the diff. |
|  `-h`, `--help`  | help |
| `--secrets string` | Use a different secrets file location (default `~/.github-keeper/secrets.yml` |


Hint: To verify the setup of all your repos use:

```shell
go run . configure-repo $(go run . list-my-repos)
```

### `completion`

Generate the autocompletion script for github-keeper for the specified shell.
See each help for each shell for details on how to use the generated script.

Usage: `github-keeper completion [shell]`

Supported Shells:
* bash
* fish
* powershell
* zsh

| Flags | Description |
|-------|-------------|
|  `-h`, `--help`  | help |

Use `github-keeper help completion [shell]` for more information about a generating completion for a specific shell.

## Installation

You can install github-keeper to `$HOME/go/bin/github-keeper` by running:

```shell
go install
```

After adding `$HOME/go/bin/` to your `PATH` you can run github-keeper by just calling `github-keeper`.

### Tips

List all repos of the integration team:

```shell
gh repo list exasol --limit 500 --json name,repositoryTopics --jq '.[] \
   | select(.repositoryTopics.[]?.name == "exasol-integration") \
   | .name' | tr "\n" " "
```

Check if a repo has disabled GitHub Actions:

```shell
gh repo list exasol --limit 500 --json name,repositoryTopics --jq '.[] \
   | select(.repositoryTopics.[]?.name == "exasol-integration") \
   | .name' | tr "\n" " " \
   | xargs go run . reactivate-scheduled-github-actions
```

## Additional Information

* [Changelog](doc/changes/changelog.md)

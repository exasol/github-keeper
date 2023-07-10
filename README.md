# GitHub Keeper

GitHub keeper is a CLI tool that helps to unify our repositories.

[![Build Status](https://github.com/exasol/github-keeper/actions/workflows/ci-build.yml/badge.svg)](https://github.com/exasol/github-keeper/actions/workflows/ci-build.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/exasol/github-keeper.svg)](https://pkg.go.dev/github.com/exasol/github-keeper)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=com.exasol%3Agithub-keeper&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=com.exasol%3Agithub-keeper)

## Obsoletion Notice

⚠ We archived this respository since the functionallity is now superseeded by the [github-issue-adapter](https://github.com/exasol/github-issue-adapter). ⚠

## Features

* Reactivate scheduled GitHub actions that have been automatically disabled by GitHub after some time.
* List all Exasol repositories that are not archived and where you are admin.
* Inspect and fix settings of a GitHub repository regarding labels, branch protection and actions.

## Usage

### Preconditions

Install Go language. Minimum version is 1.18.
* On Debian / Ubuntu:
    ```shell
    sudo apt install golang-go
    ```
* On macOS:
    ```shell
    brew install golang
    ```

### Install Without Sources

If you only want to use github-keeper, run the following command:

```shell
go install github.com/exasol/github-keeper@main
```

This will install it to `$(go env GOPATH)/bin/github-keeper`.

### Install With Sources for Development

Checkout sources and install dependent packages:

```shell
git clone https://github.com/exasol/github-keeper.git
cd github-keeper
go get ./...
```

You can install github-keeper to `$(go env GOPATH)/bin/github-keeper` by running:

```shell
go install
```

After adding `$(go env GOPATH)/bin/` to your `PATH` you can run github-keeper by just calling `github-keeper`.

## Configuration

Please create the configuration file `~/.github-keeper/secrets.yml` with the following content:

```yaml
issuesSlackWebhookUrl: "<slack web-hook url for issue updates>"
```

The url can be found in the keeper-vault.

## Usage

If you want to run github-keeper from the source code, replace the `github-keeper` command with `go run .`.

| Command                                                              | Description                                                       |
| -------------------------------------------------------------------- | ----------------------------------------------------------------- |
| `github-keeper help command`                                         | Display command line help                                         |
| `github-keeper reactivate-scheduled-github-actions`                  | Reactivate the scheduled GitHub actions for the given repository. |
| `github-keeper completion <shell>`                                   | Generate autocompletion script for shell `<shell>`                |
| `github-keeper list-my-repos`                                        | List all Exasol repositories where you are admin                  |
| `github-keeper configure-repo <repo-name> [more repo names] [flags]` | Inspect settings of GitHub repository                             |

### `list-my-repos`

List all repositories of the Exasol organization where I'm the admin and that are not archived.

Usage: `github-keeper list-my-repos [flags]`

| Flags          | Description |
| -------------- | ----------- |
| `-h`, `--help` | Help        |

### `reactivate-scheduled-github-actions`

Reenable scheduled actions automatically disabled by GitHub after some time.

Usage: `github-keeper reactivate-scheduled-github-actions <repo-name> [flags]`

| Flags          | Description |
| -------------- | ----------- |
| `-h`, `--help` | Help        |

### `configure-repo`

Verify the config of a given repository

Usage: `github-keeper configure-repo <repo-name> [more repo names] [flags]`

| Flags              | Description                                                                               |
| ------------------ | ----------------------------------------------------------------------------------------- |
| `--fix`            | If this flag is set, github-keeper fixed the findings. Otherwise it just prints the diff. |
| `-h`, `--help`     | Help                                                                                      |
| `--secrets string` | Use a different secrets file location (default `~/.github-keeper/secrets.yml`             |


Hint: To verify the setup of all your repos use:

```shell
github-keeper configure-repo $(github-keeper list-my-repos)
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

| Flags          | Description |
| -------------- | ----------- |
| `-h`, `--help` | help        |

Use `github-keeper help completion [shell]` for more information about a generating completion for a specific shell.

### Tips

List all repos of the Exasol integration team:

```shell
gh repo list exasol --limit 500 --json name,repositoryTopics --jq \
    '.[] | select(.repositoryTopics.[]?.name == "exasol-integration") | .name' \
    | tr "\n" " "
```

Reactivate disabled scheduled GitHub Actions:

```shell
gh repo list exasol --limit 500 --json name,repositoryTopics --jq \
    '.[] | select(.repositoryTopics.[]?.name == "exasol-integration") | .name' | tr "\n" " " \
    | xargs github-keeper reactivate-scheduled-github-actions
```

## Additional Information

* [Changelog](doc/changes/changelog.md)
* [Dependencies](dependencies.md)

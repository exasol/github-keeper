# GitHub Keeper

GitHub keeper is a CLI tool that helps to unify our repositories.

Features:

* List all Exasol repos where you are admin (`list-my-repos`)
* Create / Update the branch protection for the default branch according to our standards (`create-branch-protection`)

#### Installation

1. Install Go language. Minimum version is 1.13. 
2. Install dependendent packages: 
```
cd $HOME/<PATH TO THIS REPO>/product-integration-tool-chest/github-keeper
go get ./...
```

#### Usage

in the github-keeper directory run:

```shell
go run .
```

Hint: For setting up a branch protection for all your repos use:

```shell
go run . create-branch-protection $(go run . list-my-repos)
```

## Additional Information

* [Changelog](doc/changes/changelog.md)

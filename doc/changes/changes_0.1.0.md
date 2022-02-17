# Github-Keeper 0.1.0, released 2022-02-17

Code name: Improved GitHub Actions

## Features

* #1: Extracted from product-integration-tool-chest
* #6: Modified create-branch-protection to make github-action checks mandatory
* #10: Added validation for create-branch-protection
* #11: Added sonar cloud as required check to branch protection
* #18: Restricted who can merge to protected branches
* #23: Enabled delete merged branches
* #22: Enable auto-merge
* #15: Added validation for slack web-hook
* #41: Added command for reactivating disabled GitHub actions

## Refactoring

* #25: Unified commands to `configure-repo`

## Bug Fixes:

* #16: Fixed unify-labels for deprecated labels where the replacement exists
* #19: Fixed create-branch-protection for duplicated check names
* #29: Fixed configure-repo for repositories with no detected language
* #33: Fixed branch protection rule creation for projects with matrix builds
* #31: Fixed configure-repo for repositories with no workflows
* #35: Fixed branch protection creation for repos with float matrix build parameter
* #39: Fixed branch protection creation for runtime generated matrix builds (printing warning that it's not possible)
* #43: Fixed handling of directories in `.github/workflows/`

## Refactoring

* #4: Added GitHub unify-labels command (rewritten from product-integration-tool-chest)
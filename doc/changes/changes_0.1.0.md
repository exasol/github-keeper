# Github-Keeper 0.1.0, released 2021-??-??

Code name: Improved GitHub Actions

## Features:

* #1: Extracted from product-integration-tool-chest
* #6: Modified create-branch-protection to make github-action checks mandatory
* #10: Added validation for create-branch-protection
* #11: Added sonar cloud as required check to branch protection
* #18: Restricted who can merge to protected branches

## Bug Fixes:

* #16: Fixed unify-labels for deprecated labels where the replacement exists
* #19: Fixed create-branch-protection for duplicated check names

## Refactoring

* #4: Added GitHub unify-labels command (rewritten from product-integration-tool-chest)
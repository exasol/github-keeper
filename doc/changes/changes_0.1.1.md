# Github-Keeper 0.1.1, released 2022-08-??

Code name: Bugfixes

In release 0.1.1 GK now ignores workflow files that are not relevant for pull requests. You can now use floats (`3.7`),
integers (`42`) and booleans (`true`) as parameters in a matrix build. Please note that floats will always be rounded to
one decimal place, e.g. `3.74` will be rounded to `3.7` and `3.75` will be rounded to `3.8`. If this causes issues with
generated branch protection rules, please use strings (`3.75`) as parameters.

## Features:

* #50: Added validation for enable dependabot and security alerts

## Refactoring:

* #65: Added project keeper

## Bug Fixes:

* #45: Fixed validation errors for non-pullrequest workflow files
* #48: Fixed parsing of matrix parameters of types int, float and boolean
* #53: Fixed bug processing repos that do not require any branch protections
* #57: Fixed branch protection rule decision

## Dependency Updates

### Compile Dependency Updates

* Updated `golang:1.13` to `1.18`
* Updated `golang.org/x/oauth2:v0.0.0-20211104180415-d3ed0bb246c8` to `v0.6.0`
* Updated `github.com/spf13/cobra:v1.3.0` to `v1.6.1`
* Updated `gopkg.in/yaml.v3:v3.0.0-20210107192922-496545a6307b` to `v3.0.1`

### Test Dependency Updates

* Added `github.com/google/go-github/v43:v43.0.0`
* Updated `github.com/stretchr/testify:v1.7.0` to `v1.8.2`
* Removed `github.com/google/go-github/v39:v39.2.0`

### Other Dependency Updates

* Removed `gopkg.in/yaml.v2:v2.4.0`

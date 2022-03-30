# Github-Keeper 0.1.1, released 2022-02-??

Code name: Bugfixes

In release 0.1.1 GK now ignores workflow files that are not relevant for pull requests. You can now use floats (`3.7`), integers (`42`) and booleans (`true`) as parameters in a matrix build. Please note that floats will always be rounded to one decimal place, e.g. `3.74` will be rounded to `3.7` and `3.75` will be rounded to `3.8`. If this causes issues with generated branch protection rules, please use strings (`3.75`) as parameters.

## Bug Fixes:

* #45: Fixed validation errors for non-pullrequest workflow files
* #48: Fixed parsing of matrix parameters of types int, float and boolean
* #53: Fixed bug processing repos that do not require any branch protections

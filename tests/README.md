# End to End/Integration Tests for GoDBLedger

This folder describes the integration tests designed for GoDBLedger.

In this `/tests` folder there is a `/tests/evaluators` folder containing a single file for each test. The `endtoend_test.go` test will spin up a GoDBLedger instance for each test and query it as described in the evaluators file.

The `endtoend_test.go` file has been built with +build integration so will be ignored by `go test` unless the tag for "integration" is added

```
go test -tags=integration
```

alternatively calling the make target from the root directory already includes the tag

```
make test
```

To add more tests one should add a new file with the test into the evaluators directory. In the file should be an exportable Evaluator variable which is then added to the evaluators array in `endtoend_test.go`.

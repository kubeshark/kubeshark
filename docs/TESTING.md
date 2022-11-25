![Kubeshark: The API Traffic Viewer for Kubernetes](https://raw.githubusercontent.com/kubeshark/assets/master/svg/kubeshark-logo.svg)
# Testing guidelines

## Generic guidelines
* Use "[testing](https://pkg.go.dev/testing)" package
* Write [Table-driven tests using subtests](https://go.dev/blog/subtests)
* Use cleanup in test/subtest in order to clean up resources
* Name the test func "Test<tested_func_name><tested_case>"

## Unit tests
* Position the test file inside the folder of the tested package
* In case of internal func testing
  * Name the test file "<tested_file_name>_internal_test.go"
  * Name the test package same as the package being tested
  * Example - [Config](../cli/config/config_internal_test.go)
* In case of exported func testing
  * Name the test file "<tested_file_name>_test.go"
  * Name the test package "<tested_package>_test"
  * Example - [Slice Utils](../cli/kubeshark/sliceUtils_test.go)
* Make sure to run test coverage to make sure you covered all the cases and lines in the func

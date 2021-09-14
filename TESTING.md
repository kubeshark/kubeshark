![Mizu: The API Traffic Viewer for Kubernetes](assets/mizu-logo.svg)
# TESTING
Testing guidelines for Mizu project

## Generic guidelines
* Use testing package
* Write Table-driven tests using subtests (https://go.dev/blog/subtests)
* Use cleanup in test/subtest in order to clean up resources
* Test func name - Test<case_being_tested>

## Unit tests
* Test file position - Inside the folder of the tested package
* In case of internal func testing
  * Test file name - <tested_file_name>_internal_test.go
  * Test package name - same as the package being tested
  * Example - [Config](cli/config/config_internal_test.go)
* In case of exported func testing
  * Test file name - <tested_file_name>_test.go
  * Test package name - <package_being_tested>_test
  * Example - [Slice Utils](cli/mizu/sliceUtils_test.go)
  
## Acceptance tests
* Test file position - Inside the acceptance tests folder
* File name - <tested_command>_test.go
* Package name - acceptanceTests
* Add short check and skip
* Use/Create generic tests func in acceptanceTests/testsUtils
* Don't use sleep inside the tests - active check 
* Running acceptance tests locally
  * Switch to the branch that is being tested
  * Run acceptanceTests/setup.sh
  * Run tests
* Example - [Tap](acceptanceTests/tap_test.go)  
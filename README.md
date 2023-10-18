# Single Pass Private Information Retrieval

Single Pass PIR is a new PIR protocol with extremely fast preprocessing and query time. This repo is a fork from the Checklist repository. In order to run our tests, please go into the pir directory and run:

go test 

To single out any specific test, you can comment out the check for t.Short within pir_test.go and add the additional -short flag.

The number of iterations each test runs for was reset to 1 so that the tests run faster, however, this is tunable with numIter to get an average over more repetitions.


The bulk of our code is within ./pir/pir_single_pass.go, ./psetggm/permute.cpp, ./psetggm/xor/cpp
./psetggm/answer.cpp, ./pir/dynamic_db.go



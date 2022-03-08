#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Echo commands to console
set -x

# Wait until db is ready for testing
/usr/local/bin/whenavail db 3306 100 echo "database is ready for tests"

# Tests need to be run sequentially for database access and fixture integrity
testDirs=`find -name '*_test.go' -not -path "*vendor*" -printf '%h\n' | sort -u`
i=0
for testDir in $testDirs; do
    go test $testDir
done

# Print the Serverless version in the logs
serverless --version

#!/bin/bash
test -e ssshtest || wget -q https://raw.githubusercontent.com/ryanlayer/ssshtest/master/ssshtest

. ssshtest

PARENT_DIR=`git rev-parse --show-toplevel`
export PATH="${PATH}:${PARENT_DIR}"

set -o nounset

# Main
run test_g gg
assert_no_stderr

# Logout
run test_logout gg logout
assert_in_stderr "Successfully Logged out"

# No results query
run test_ls gg ls
assert_no_stdout

# Login
run test_login gg sync --token ${TEST_TOKEN}
assert_in_stderr "ggtest-2"

# Delete leftover gists

# Create new gist - stdin
run test_new gg new db.go
assert_in_stderr https://gist.github.com/


# Delete a gist
run rm_gist gg rm 0
assert_in_stderr "Removed"
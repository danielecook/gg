#!/usr/bin/bash
test -e ssshtest || wget -q https://raw.githubusercontent.com/ryanlayer/ssshtest/master/ssshtest

. ssshtest

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

stdin_test() {
    cat db.go | gg new --description "test -- stdin"
}

# Create new gist - stdin
run test_new_stdin stdin_test
assert_in_stdout https://gist.github.com/

# Create new gist - filename
run test_new_files gg new --description "test -- files" db.go README.md
assert_in_stdout https://gist.github.com/

# Delete a gist
run rm_gist gg rm 0
assert_in_stderr "Removed"
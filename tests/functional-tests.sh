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


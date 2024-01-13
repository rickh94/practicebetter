#!/usr/bin/env bash
exec 1>&2 # redirect output to stderr
if git diff --cached | grep [dD]O-NOT-COMMIT; then
    echo "[D]O-NOT-COMMIT found - rejecting commit"
    exit 1
fi

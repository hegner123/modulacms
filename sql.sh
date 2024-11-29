#!/usr/bin/env bash


if [[ "$1" == "user" ]]; then
    cat sql/scripts/list_all_users.sql | sqlite3 modula_test.db
fi

#!/usr/bin/env bash

grep GenerateUUID consul/state/state_store.go
RESULT=$?
if [ $RESULT -eq 0 ]; then
    exit 1
fi

grep GenerateUUID consul/fsm.go
RESULT=$?
if [ $RESULT -eq 0 ]; then
    exit 1
fi

#!/bin/bash

green() {
  printf "\e[32m%s\e[0m\n" "$1"
}

echo "Running unit tests..."
unit_test_result=$(go test -tags=unit ./... -v)
echo "$unit_test_result"

if [[ $? -ne 0 ]]; then
    echo "Unit tests failed."
    exit 1
fi

echo "Running integration tests..."
integration_test_result=$(go test -tags=integration ./... -v)
echo "$integration_test_result"

if [[ $? -ne 0 ]]; then
    echo "Integration tests failed."
    exit 1
fi

green "All tests (unit and integration) passed successfully!"

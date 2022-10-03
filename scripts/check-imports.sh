#!/bin/bash

readonly SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
readonly REPO_ROOT="${SCRIPT_DIR}/.."
readonly GOIMPORTS="${REPO_ROOT}/bin/goimports"

if [[ -n "$($GOIMPORTS --local github.com/giantswarm/fleet-membership-operator-gcp -l .)" ]]; then
  $GOIMPORTS --local github.com/giantswarm/fleet-membership-operator-gcp -d .
  exit 1
fi

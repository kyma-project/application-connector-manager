#!/usr/bin/env bash

# JOB_NAME      - the name of the job that is runing tests (required)
# LOGS_OUT      - path to directory where the logs will be stored (optional, default: ${PWD})
# NAMESPACE the - namespace where the test job is runing (optional,default: test)
# TEST_TIMEOUT  - duration to wait on tests to finish (optional, default: 900s)

function fetch_tests() {
  local JOB_NAME=$1
  # check if JOB_NAME is provided
  if [ -z "$JOB_NAME" ]; then
    echo "Usage: $0 <JOB_NAME> [LOGS_OUT] [NAMESPACE] [TEST_TIMEOUT]"
    exit 1
  fi
  local LOGS_OUT=${2:-${PWD}}
  local NAMESPACE=${3:-test}
  local TEST_TIMEOUT=${4:-30m}
  # wait for the job to finish
  kubectl wait job/$JOB_NAME \
	-n $NAMESPACE \
	--for=condition=complete \
	--timeout=$TEST_TIMEOUT
  # store the exit code of the job
  local __job_result__=$?
  # try to get the logs of the job and store them in the TEST_LOG file
  kubectl logs \
	  -n $NAMESPACE \
	  -f job/$JOB_NAME \
	  2>&1 > "$LOGS_OUT"/$JOB_NAME.log
  # exit with original job exit code
  exit $__job_result__
}

fetch_tests $@

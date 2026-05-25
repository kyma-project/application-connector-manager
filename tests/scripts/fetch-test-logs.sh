#!/usr/bin/env bash

# JOB_NAME      - the name of the job that is runing tests (required)
# LOGS_OUT      - path to directory where the logs will be stored (optional, default: ${PWD})
# NAMESPACE     - the namespace where the test job is runing (optional, default: test)
# TEST_TIMEOUT  - duration to wait on tests to finish (optional, default: 300s)

function fetch_tests() {
  local JOB_NAME=$1
  if [ -z "$JOB_NAME" ]; then
    echo "Usage: $0 <JOB_NAME> [LOGS_OUT] [NAMESPACE] [TEST_TIMEOUT]"
    exit 1
  fi
  local LOGS_OUT=${2:-${PWD}}
  local NAMESPACE=${3:-test}
  local TEST_TIMEOUT=${4:-300s}

  # Wait for the job to either complete or fail, whichever comes first.
  # `kubectl wait --for=condition=complete` alone blocks until the timeout
  # when the job fails, so we race it against `condition=failed`.
  kubectl wait "job/$JOB_NAME" -n "$NAMESPACE" \
    --for=condition=complete --timeout="$TEST_TIMEOUT" &
  local complete_pid=$!
  kubectl wait "job/$JOB_NAME" -n "$NAMESPACE" \
    --for=condition=failed --timeout="$TEST_TIMEOUT" \
    && kill "$complete_pid" 2>/dev/null &
  local failed_pid=$!

  wait -n "$complete_pid" "$failed_pid"
  local __job_result__=$?

  # Reap the loser so we don't leak a background process.
  kill "$complete_pid" "$failed_pid" 2>/dev/null
  wait "$complete_pid" "$failed_pid" 2>/dev/null

  # Capture logs from all pods of the job. No `-f` — the pod has terminated
  # by now and `-f` is unreliable against finished/failed pods.
  kubectl logs -n "$NAMESPACE" --tail=-1 --all-containers --prefix \
    "job/$JOB_NAME" > "$LOGS_OUT/$JOB_NAME.txt" 2>&1 || true

  # Echo logs to stdout so CI surfaces them on failure.
  echo "----- logs from job/$JOB_NAME -----"
  cat "$LOGS_OUT/$JOB_NAME.txt"
  echo "----- end logs -----"

  # On failure, also dump pod state for debugging.
  if [ "$__job_result__" -ne 0 ]; then
    kubectl get pods -n "$NAMESPACE" -l "job-name=$JOB_NAME" -o wide || true
    kubectl describe pods -n "$NAMESPACE" -l "job-name=$JOB_NAME" || true
  fi

  exit $__job_result__
}

fetch_tests "$@"

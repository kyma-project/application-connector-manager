#!/usr/bin/env bash

# JOB_NAME      - the name of the job that is runing tests (required)
# LOGS_OUT      - path to directory where the logs will be stored (optional, default: ${PWD})
# NAMESPACE     - the namespace where the test job is runing (optional, default: test)
# TEST_TIMEOUT  - duration to wait on tests to finish, accepts Ns/Nm/N (optional, default: 300s)

function to_seconds() {
  local v=$1
  case "$v" in
    *s) echo "${v%s}" ;;
    *m) echo $(( ${v%m} * 60 )) ;;
    *h) echo $(( ${v%h} * 3600 )) ;;
    *)  echo "$v" ;;
  esac
}

function fetch_tests() {
  local JOB_NAME=$1
  if [ -z "$JOB_NAME" ]; then
    echo "Usage: $0 <JOB_NAME> [LOGS_OUT] [NAMESPACE] [TEST_TIMEOUT]"
    exit 1
  fi
  local LOGS_OUT=${2:-${PWD}}
  local NAMESPACE=${3:-test}
  local TEST_TIMEOUT_RAW=${4:-300s}
  local TEST_TIMEOUT_S
  TEST_TIMEOUT_S=$(to_seconds "$TEST_TIMEOUT_RAW")

  # Poll the job's terminal conditions directly. Racing two `kubectl wait`
  # commands via `wait -n` is unreliable: when the job fails, the failed-wait
  # exits 0 *and* its chained `kill` exits 0, so the shell can return 0 for
  # the failed job and mark a real failure as success.
  local deadline=$(( $(date +%s) + TEST_TIMEOUT_S ))
  local job_result=
  while [ "$(date +%s)" -lt "$deadline" ]; do
    local complete failed
    complete=$(kubectl get job "$JOB_NAME" -n "$NAMESPACE" \
      -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' 2>/dev/null)
    failed=$(kubectl get job "$JOB_NAME" -n "$NAMESPACE" \
      -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null)
    if [ "$complete" = "True" ]; then
      job_result=0
      break
    fi
    if [ "$failed" = "True" ]; then
      job_result=1
      break
    fi
    sleep 5
  done
  if [ -z "$job_result" ]; then
    echo "timed out after ${TEST_TIMEOUT_S}s waiting for job/$JOB_NAME to reach a terminal condition"
    job_result=1
  fi

  # Capture logs per-pod via selector. `kubectl logs job/X` is unreliable —
  # if the picked pod has been GC'd or hasn't reached a loggable state,
  # kubectl emits "timed out waiting for the condition" instead of logs.
  local logs_file="$LOGS_OUT/$JOB_NAME.txt"
  : > "$logs_file"
  local pods
  pods=$(kubectl get pods -n "$NAMESPACE" -l "job-name=$JOB_NAME" \
    -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
  if [ -z "$pods" ]; then
    echo "no pods found for job/$JOB_NAME (may have been garbage collected)" >> "$logs_file"
  else
    for pod in $pods; do
      echo "----- pod $pod -----" >> "$logs_file"
      kubectl logs -n "$NAMESPACE" --tail=-1 --all-containers --prefix "$pod" \
        >> "$logs_file" 2>&1 || true
    done
  fi

  echo "----- logs from job/$JOB_NAME -----"
  cat "$logs_file"
  echo "----- end logs -----"

  if [ "$job_result" -ne 0 ]; then
    kubectl get pods -n "$NAMESPACE" -l "job-name=$JOB_NAME" -o wide || true
    kubectl describe pods -n "$NAMESPACE" -l "job-name=$JOB_NAME" || true
  fi

  exit "$job_result"
}

fetch_tests "$@"

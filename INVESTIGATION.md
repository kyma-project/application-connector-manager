# Validator Test Investigation

## Problem Statement
`make -C tests/hack/ci k3d-validator-tests` fails with "error: timed out waiting for the condition on jobs/application-connectivity-validator-test".

The test Job runs tests that make HTTP requests to `central-application-connectivity-validator.kyma-system:8080`. All requests fail with `net.OpError`.

---

## Lead 1: Validator CrashLoopBackOff (PARTIALLY FIXED)

**Observation:** Validator pods crash with `FATAL: failed to wait for application caches to sync kind source: *v1alpha1.Application: timed out waiting for cache to be synced`.

**Root cause:** The validator uses controller-runtime to watch Application CRs. During cache sync, it calls the Kubernetes API server. The API server connection goes through the Istio sidecar. In the early cluster lifecycle (right after Calico CNI is installed), the pods on `k3d-kyma-server-0` have intermittent network issues — the sidecar can't proxy connections to the API server (TLS handshake timeout).

**Evidence:**
- Logs show: `Get "https://10.43.0.1:443/api": net/http: TLS handshake timeout`
- Manually deleting the pods and letting them reschedule fixes the crash
- The manager pod (same NetworkPolicies) works because it lands on a different node or starts at a luckier time
- Both pods eventually recover after 2-4 restarts (backoff gives Calico time to stabilize)

**Fix applied:** Added Calico readiness waits in CI Makefile + wait for validator rollout before running tests.

**Status:** PARTIALLY EFFECTIVE — reduces restart count but doesn't eliminate the race entirely. The validator still restarts 2-4 times before stabilizing.

---

## Lead 2: Test Job Namespace Mismatch (FIXED)

**Observation:** Test Job was deployed to `istio-system` instead of `test`.

**Root cause:** Commit `687ad4f` changed `values.yaml` from `global.namespace: "test"` to `global.namespace: "istio-system"`. The `fetch-test-logs.sh` default was also `istio-system`.

**Evidence:**
- `kubectl get jobs -A | grep validator` showed job in `istio-system`
- `helm template` with `--set global.namespace=test` correctly renders to `test` (Helm `--set` overrides `--values`)

**Fix applied:** Reverted `values.yaml` to `namespace: "test"`, reverted `fetch-test-logs.sh` default.

**Status:** FIXED — Job now deploys to correct namespace `test`.

---

## Lead 3: NetworkPolicy Source Namespace Selector Fails Through ClusterIP (CURRENT BLOCKER)

**Observation:** Test pod in namespace `test` CANNOT reach the validator service ClusterIP on port 8080, despite NetworkPolicy explicitly allowing ingress from namespace `test`.

**Root cause:** Calico with k3s service proxy (not standard kube-proxy) loses source namespace identity when traffic is routed through a ClusterIP service. The kube-proxy equivalent performs DNAT, and Calico's policy enforcement can't map the post-DNAT packet back to the source pod's namespace.

**Evidence:**
- Direct pod IP (`192.168.194.72:8080`) from test namespace: **WORKS**
- ClusterIP (`10.43.49.58:8080`) from test namespace: **FAILS** (timeout)
- ClusterIP port 8081 (health check, no `from` restriction in policy): **WORKS**
- After removing `namespaceSelector` from the policy (allow from anywhere): **WORKS** (verified with busybox pod without sidecar)

**Fix applied:** Removed `namespaceSelector` from `validator-test-allow-to-validator` policy — now allows port 8080 from anywhere (acceptable for test environment).

**Status:** PARTIALLY FIXED — direct connectivity test works, but the test pod with Istio sidecar still fails. See Lead 4.

---

## Lead 4: Test Pod Sidecar Prevents Connectivity (FIXED)

**Observation:** Even with the NetworkPolicy fix, the test Job still fails. The test pod has a sidecar injected (namespace has `istio-injection=enabled`). The annotation `excludeOutboundPorts: "8080"` should bypass the sidecar for port 8080 traffic.

**Evidence:**
- Busybox pod WITHOUT sidecar (`sidecar.istio.io/inject: "false"`) in namespace `test` can reach validator ClusterIP:8080 ✓
- Test pod WITH sidecar and `excludeOutboundPorts: "8080"` fails with `net.OpError`
- Error timings are 0.00s and 1.02s (connection refused/reset, NOT timeout)
- `localhost:15000/quitquitquit` and `localhost:15020/quitquitquit` in TearDown also fail — confirms sidecar is NOT running

**Fix applied:** Added `sidecar.istio.io/inject: "false"` to the test Job pod template. Removed `excludeOutboundPorts` (unnecessary without sidecar).

**Status:** FIXED — test pod now connects to the validator successfully. Tests run and produce HTTP responses (not `net.OpError`).

---

## Lead 5: Validator Sidecar Blocking Inbound Despite excludeInboundPorts (ABANDONED)

**Observation:** Validator has `traffic.sidecar.istio.io/excludeInboundPorts: "8080,15020"`.

**Why abandoned:** Direct pod IP access from test namespace works (Lead 3 evidence), proving the validator IS listening on 8080 and accepting connections. The sidecar correctly excludes port 8080 from interception. The issue is in the routing TO the pod, not the pod's ability to accept.

---

## Lead 6: API Server Port 443 vs 6443 DNAT Issue (ABANDONED)

**Observation:** The NetworkPolicy `acm-module-to-api-server` allows egress on port 443. The actual API server endpoint is `172.20.0.3:6443`. After kube-proxy DNAT, the destination changes from 443 to 6443.

**Why abandoned:** The manager pod with the SAME policies works fine. The Istio sidecar handles the API server connection by connecting directly to the endpoint (bypassing kube-proxy DNAT through its own service discovery). Pods with sidecar work; pods without don't. Since both the manager and validator have sidecars, this isn't the differentiator — the validator's issue is timing (sidecar not ready when first connection attempt is made).

---

## Lead 7: Application CR Cache Not Synced When Test Starts (FIXED)

**Observation:** After fixing Lead 4, the test pod connects to the validator but gets 404: `"application data for name event-test-standalone is not found in the cache. Please retry"`.

**Root cause:** The Application CRs (`event-test-compass`, `event-test-standalone`) are applied at the same time as the test Job. The validator uses a controller-runtime informer to watch Application CRs. The informer needs time to receive the watch event and populate the cache. The test Job starts immediately and queries the validator before the cache is populated.

**Evidence:**
- Manual curl test AFTER the validator has been running for a few minutes returns HTTP 500 (app IS in cache, cert header missing) — confirms cache syncs eventually
- Test Job starts within 2-3s of Application CR creation — not enough time for the watch event

**Fix applied:** 
1. Application CRs are now applied as a separate Makefile step BEFORE the test Job
2. Added an init container `wait-for-cache` to the test Job that polls the validator until it returns non-404/non-000 (meaning the cache has synced)

**Status:** FIXED — init container correctly waits for cache sync. Confirmed init container logs show "Cache ready (HTTP 500)".

---

## Lead 8: Validator CrashLoopBackOff Persists (CURRENT BLOCKER)

**Observation:** The validator pods keep crashing with "failed to wait for application caches to sync kind source: *v1alpha1.Application: timed out waiting for cache to be synced". Even after Calico stabilizes, the pods continue to cycle in CrashLoopBackOff (7+ restarts, exponential backoff of 2-5 minutes between attempts).

**Root cause:** The validator uses controller-runtime to start a `Kind` source informer that watches Application CRs. This requires a working connection to the kube-apiserver. The connection goes through the Istio sidecar proxy. The sidecar itself needs to connect to the API server (for certificate validation via Pilot). In the early k3d cluster lifecycle:

1. Calico Felix may still be programming iptables rules
2. The Istio sidecar's connection to istiod may be unstable
3. The API server's endpoints may not be fully reachable from all nodes

When the informer Start() call times out (controller-runtime default: 2min), the validator logs FATAL and exits. Kubernetes restarts it, but with exponential backoff the wait grows from 10s → 20s → 40s → 80s → 160s → 320s → 5min.

**The sequence:**
1. `kubectl rollout status` succeeds (validator passes health check on :8081 before cache sync completes)
2. Background cache sync for Application CRs times out
3. Validator exits with FATAL
4. Pod enters CrashLoopBackOff
5. Eventually (after 4-6 restarts, ~3-5 minutes), cluster stabilizes and validator successfully syncs

**Evidence:**
- Validator logs: `Get "https://10.43.0.1:443/api": net/http: TLS handshake timeout`
- Followed by: `FATAL: failed to wait for application caches to sync kind source: *v1alpha1.Application`
- Both replicas exhibit same behavior
- After enough restarts (7+), both pods eventually become stable (2/2 Running)
- `rollout status` is unreliable because the readiness probe on :8081 succeeds BEFORE the cache sync finishes

**Why previous fix (Lead 1) is insufficient:**
- Calico readiness waits (`kubectl wait tigerastatus/calico`) confirm the Calico control plane is ready
- But individual nodes may still have incomplete dataplane programming (iptables rules)
- The validator starts on nodes that haven't fully converged, causing the TLS timeout

**Current approach:**
- The Makefile now has a stability loop: wait for `readyReplicas == replicas` and confirm it stays stable for 10 seconds
- This works but can take 3-5 minutes of waiting after deployment
- The init container on the test Job provides a second safety net

**Possible fixes:**
1. **Increase backoffLimit** on the test Job from 3 to 6+ (allows more retries while validator stabilizes)
2. **Increase the stability wait timeout** in Makefile (wait up to 10 minutes for validator)
3. **Fix the validator code** to retry cache sync instead of fatally exiting (proper fix, but out of scope for CI tests)
4. **Delete validator pods after Calico stabilizes** to force immediate restart without backoff (hack but effective)

### Recommendation: Option 4 (delete pods) + Option 1 (increase backoffLimit)

Deleting the validator pods right after confirming Calico is stable forces an immediate restart (no backoff delay). The fresh pods start on a now-stable cluster and sync successfully on first try. Increasing the test Job's backoffLimit provides additional resilience.

---

## Current Ideas to Fix

### Option A: Disable Istio injection for the test pod ✓ APPLIED
Add `sidecar.istio.io/inject: "false"` annotation to the test Job pod template. The test doesn't need mTLS — it's explicitly testing plain HTTP to port 8080 (which is excluded from sidecar on both sides anyway).

### Option B: Add `holdApplicationUntilProxyStarts: true` annotation — NOT NEEDED
This tells Istio to block the app container startup until the sidecar is ready. But it may cause issues with the Job (sidecar won't terminate and the Job won't complete). Moot point since we disabled sidecar injection (Option A).

### Option C: Add an init container to the test Job that waits for connectivity ✓ APPLIED
An init container that loops until the validator responds with a non-404/non-000 HTTP code before starting the tests. Confirms both connectivity AND cache readiness.

### Option D: Delete validator pods after Calico stabilizes — NEXT TO TRY
Add `kubectl delete pods -n kyma-system -l app=central-application-connectivity-validator` after Calico readiness waits. This forces fresh pod creation without exponential backoff, on a now-stable cluster. Combined with the Makefile stability wait, should result in validator coming up clean on first attempt.

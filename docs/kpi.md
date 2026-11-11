# Application Connector Module — KPI Framework

> **Version:** 1.0  
> **Date:** 2026-11-11  
> **Stakeholders:** Team, Varban, Artur  
> **Review cadence:** Monthly

---

## Purpose

This document defines the Key Performance Indicators (KPIs) for the Application Connector Module (ACM). It covers:

- What each KPI measures and how it is collected
- Why it was chosen and what business value it reflects
- Thresholds (warning and critical) and the action plan for each breach

The KPIs are designed to surface technical performance, reliability, scalability, and product health in a form that is meaningful both to the engineering team and to management stakeholders.

---

## Table of Contents

1. [KPI Overview](#1-kpi-overview)
2. [Reliability KPIs](#2-reliability-kpis)
   - [2.1 Module Availability](#21-module-availability)
   - [2.2 Successful UCL Sync Rate](#22-successful-ucl-sync-rate)
   - [2.3 Certificate Renewal Risk](#23-certificate-renewal-risk)
3. [Performance KPIs](#3-performance-kpis)
   - [3.1 Application Gateway Latency (p99)](#31-application-gateway-latency-p99)
   - [3.2 Application Gateway Error Rate](#32-application-gateway-error-rate)
   - [3.3 Application Connectivity Validator Error Rate](#33-application-connectivity-validator-error-rate)
4. [Scalability KPIs](#4-scalability-kpis)
   - [4.1 Number of Connected External Systems per Runtime](#41-number-of-connected-external-systems-per-runtime)
   - [4.2 Request Throughput Capacity](#42-request-throughput-capacity)
5. [Product Health KPIs](#5-product-health-kpis)
   - [5.1 Operator Reconciliation Success Rate](#51-operator-reconciliation-success-rate)
   - [5.2 Module Installation / Upgrade Success Rate](#52-module-installation--upgrade-success-rate)
   - [5.3 Time to First Successful Integration](#53-time-to-first-successful-integration)
6. [Code Quality & Maintenance KPIs](#6-code-quality--maintenance-kpis)
   - [6.1 Test Coverage](#61-test-coverage)
   - [6.2 Critical Open Bug Age](#62-critical-open-bug-age)
   - [6.3 Mean Time to Recovery (MTTR)](#63-mean-time-to-recovery-mttr)
7. [Measurement Infrastructure](#7-measurement-infrastructure)
8. [Review Process and Stakeholder Alignment](#8-review-process-and-stakeholder-alignment)

---

## 1. KPI Overview

| # | KPI | Category | Target | Warning | Critical |
|---|-----|----------|--------|---------|----------|
| 2.1 | Module Availability | Reliability | ≥ 99.9 % | < 99.9 % | < 99.5 % |
| 2.2 | UCL Sync Success Rate | Reliability | ≥ 99 % | < 99 % | < 95 % |
| 2.3 | Certificate Renewal Risk | Reliability | 0 at-risk certs | any cert < 14d without renewal | any cert < 7d without renewal |
| 3.1 | Gateway Latency p99 | Performance | ≤ 500 ms | > 500 ms | > 1 000 ms |
| 3.2 | Gateway Error Rate | Performance | ≤ 0.5 % | > 0.5 % | > 2 % |
| 3.3 | Validator Error Rate | Performance | ≤ 0.1 % | > 0.1 % | > 1 % |
| 4.1 | Connected Systems per Runtime | Scalability | ≥ current baseline | — | regression |
| 4.2 | Request Throughput Capacity | Scalability | ≥ baseline | < 80 % baseline (load test) | < 60 % baseline (load test) |
| 5.1 | Reconciliation Success Rate | Product Health | ≥ 99 % | < 99 % | < 95 % |
| 5.2 | Install / Upgrade Success Rate | Product Health | ≥ 99 % | < 99 % | < 95 % |
| 5.3 | Time to First Integration | Product Health | ≤ 10 min | > 10 min | > 20 min |
| 6.1 | Test Coverage | Quality | ≥ 80 % | < 80 % | < 70 % |
| 6.2 | Critical Bug Age | Quality | ≤ 5 days | > 5 days | > 10 days |
| 6.3 | MTTR | Quality | ≤ 4 h | > 4 h | > 8 h |

---

## 2. Reliability KPIs

### 2.1 Module Availability

**Definition:** Percentage of time all three core ACM components — Application Gateway (CAG), Application Connectivity Validator (CACV), and Compass Runtime Agent (CRA) — have their target replica count available.

**How it is measured:**
- Source: Kubernetes Deployment `status.availableReplicas` vs. `spec.replicas`.
- Aggregated over a rolling 30-day window.
- Collected by a Prometheus recording rule; exposed on the team dashboard.

**Why it was chosen:**
ACM sits on the critical path for every inbound event from external systems and every outbound API call from Kyma workloads. Any period of partial or full unavailability directly blocks customer extensions. This KPI is the single most visible indicator of whether the module is serving its purpose.

**Thresholds and action plan:**

**Warning** — < 99.9 % in a 30-day window:
1. Identify the affected component and the time window.
2. Check pod events and logs for OOMKills, CrashLoops, or evictions.
3. Review recent deployments or config changes.
4. Adjust resource limits or HPA settings if needed.
5. Notify the team at the next standup.

**Critical** — < 99.5 % in a 30-day window, or any 5-minute outage of all replicas of a component:
1. Declare incident.
2. Page on-call engineer.
3. Roll back the last deployment or config change if applicable.
4. Scale up replicas manually if autoscaler is stuck.
5. Post incident summary within 48 business hours.
6. Create follow-up task to prevent recurrence.


---

### 2.2 Successful UCL Sync Rate

**Definition:** Percentage of Runtime Agent synchronization cycles that complete without error.

**How it is measured:**
- Source: Runtime Agent metrics (sync attempt counter vs. error counter) or CompassConnection CR status conditions.
- Measured over a rolling 24-hour window.
- A sync cycle is considered failed if the CompassConnection CR reaches `SynchronizationFailed` status.

**Why it was chosen:**
The Runtime Agent is responsible for automatically keeping external system registrations in sync with UCL. A high failure rate means new systems are not onboarded automatically, registered APIs are stale, or certificate data is not updated — all of which silently break customer integrations. This KPI measures the health of the automated integration path that most customers rely on.

**Thresholds and action plan:**

**Warning** — < 99 % over 24 h:
1. Check CompassConnection CR status and events.
2. Inspect CRA logs for network errors, auth failures, or UCL API timeouts.
3. Verify UCL Director is reachable from the Kyma cluster.
4. Alert team; monitor for improvement within 1 business hour.

**Critical** — < 95 % over 24 h, or 3 or more consecutive sync failures:
1. Declare incident.
2. Check UCL Director status (is the issue on ACM's side or UCL's side?).
3. If UCL is healthy: collect CRA heap dump and restart the pod.
4. If UCL is degraded: track UCL's incident.
5. Post root-cause analysis within 48 business hours.

---

### 2.3 Certificate Renewal Risk

**Definition:** Number of managed certificates approaching expiry without a successful Runtime Agent renewal.

**How it is measured:**
- Source: CompassConnection CR status field `CertificateStatus.NotAfter` and renewal log events from CRA.
- Alerting rule: trigger if a certificate with `NotAfter` within 7 days has not had a successful renewal attempt in the last 24 h.
- Note: because `NotAfter` and log events do not expose a durable attempt/result counter, this KPI is expressed as a risk count (number of at-risk certificates) rather than a percentage success rate.

**Why it was chosen:**
ACM uses mTLS as the primary security mechanism for trusted communication between external systems and Kyma. A missed certificate renewal causes the mTLS handshake to fail, which breaks all inbound event delivery. Because this failure is time-triggered and silent until expiry, it requires proactive monitoring. A failure here is almost always a P1 incident.

**Thresholds and action plan:**

**Warning** — any certificate with fewer than 14 days validity and no successful renewal in 48 h:
1. Check CRA logs for renewal errors.
2. Verify UCL Connector is reachable.
3. Manually trigger a reconnect if the automated path is stuck (see tutorial: reconnect-runtime-agent-with-compass).

**Critical** — any certificate with fewer than 7 days validity and no successful renewal in 24 h:
1. Declare incident immediately.
2. Manually rotate certificate following the revocation and reconnect runbook.
3. Root-cause analysis required within 24 business hours of resolution.

---

## 3. Performance KPIs

### 3.1 Application Gateway Latency (p99)

**Definition:** The 99th-percentile end-to-end proxied request latency observed by Application Gateway, measured from request receipt at the gateway until the response is forwarded back to the Kyma workload. This includes network and external target-system latency.

**How it is measured:**
- Source: Application Gateway Prometheus histogram `http_request_duration_seconds` (or equivalent).
- Measured over a 5-minute rolling window.
- Reported as daily p99 on the dashboard; alert on a 5-minute breach.

**Why it was chosen:**
Application Gateway sits on every outbound call from a Kyma workload to an external API. High proxied latency degrades the responsiveness of every customer extension that calls external systems (S/4HANA, BTP services, third-party APIs). The p99 is chosen over average because tail latency directly affects the worst-case user experience. When the p99 spikes, filtering by `app_name` label will reveal whether the cause is a slow external target or a gateway-internal issue (e.g. token-cache miss, connection pool exhaustion).

**Thresholds and action plan:**

**Warning** — p99 > 500 ms sustained for 5 min:
1. Check if a specific application or target system is causing the tail latency (filter by `app_name` label).
2. Inspect OAuth token cache hit rate — a miss forces an extra round-trip.
3. Review HPA state: is the gateway under CPU/memory pressure?

**Critical** — p99 > 1 000 ms sustained for 5 min:
1. Declare incident.
2. Immediately check if a single slow target is dragging up the p99 — if so, consider rate-limiting or isolating it.
3. Scale up gateway replicas.
4. If a recent deployment is correlated, roll back.
5. Post incident summary within 48 business hours.

---

### 3.2 Application Gateway Error Rate

**Definition:** Percentage of proxied requests that result in a 5xx response code returned to the Kyma workload, excluding cases where the external target system itself responded with 5xx (which are rewritten to 502 with a `Target-System-Status` header).

**How it is measured:**
- Source: Application Gateway metrics counter, distinguishing ACM-origin 5xx from target-origin 5xx (via `Target-System-Status` header presence).
- Measured over a rolling 5-minute window.

**Why it was chosen:**
Errors from Application Gateway break outbound API calls from Kyma workloads. Unlike latency, errors have immediate functional impact — a failed call means the customer's extension logic did not execute. Distinguishing ACM-origin errors from target errors is critical so the team knows whether to act or to communicate to the customer that the issue is on the external system side.

**Thresholds and action plan:**

**Warning** — ACM-origin error rate > 0.5 % over 5 min:
1. Check gateway logs for auth failures (OAuth, mTLS), DNS resolution errors, or connection timeouts.
2. Verify the Application CR for the affected app has correct credentials.
3. Alert team.

**Critical** — ACM-origin error rate > 2 % over 5 min:
1. Declare incident.
2. Determine if errors are global (all apps) or isolated (one app/target).
3. If global: check gateway pod health, restart if in bad state.
4. If isolated: investigate the specific Application CR and its credentials secret.

---

### 3.3 Application Connectivity Validator Error Rate

**Definition:** Percentage of inbound requests from external systems rejected by the Application Connectivity Validator due to certificate validation failure.

**How it is measured:**
- Source: Connectivity Validator metrics or Istio access logs filtered to rejection responses from the validator.
- Measured over a rolling 5-minute window.
- Expected baseline: near zero under normal operation (valid external systems always present valid certs).

**Why it was chosen:**
The validator is the security gatekeeper for all inbound event traffic. A high rejection rate indicates either a security attack (spoofed requests), a misconfiguration, or a certificate expiry/mismatch on the external system side. Because false rejections block legitimate event delivery, even a small unexpected increase warrants investigation.

**Thresholds and action plan:**

**Warning** — rejection rate > 0.1 % over 5 min:
1. Check validator logs for the certificate subject being rejected.
2. Correlate with UCL sync status — a sync failure may have caused a stale certificate.

**Critical** — rejection rate > 1 % over 5 min, or a known-good external system is consistently rejected:
1. Declare incident.
2. Inspect the Certificate Secret in the cluster — has the certificate been updated or corrupted?
3. If a UCL sync recently ran, check if it overwrote the correct certificate with an incorrect one.
4. If the certificate looks correct, check Istio Gateway configuration for mTLS policy changes.

---

## 4. Scalability KPIs

### 4.1 Number of Connected External Systems per Runtime

**Definition:** The number of Application CRs successfully managed per Kyma runtime instance. Tracked as a distribution across all managed runtimes, not just a single value.

**How it is measured:**
- Source: Count of Application CRs in each cluster, collected via Kubernetes API or a Prometheus gauge.
- Reported as: median, p90, and maximum across all observed runtimes. Tracked over time to detect trend changes.

**Why it was chosen:**
One of ACM's core value propositions is that it scales to support many external systems per runtime without operator intervention. Tracking this KPI serves two purposes: (1) it validates that the module meets customer demand as adoption grows, and (2) regression in this metric (e.g. Applications failing to be created or being unexpectedly deleted) is an early signal of a Runtime Agent or UCL sync issue.

**Thresholds and action plan:**

**Warning** — p90 count decreases by > 10 % month-over-month without a known cause:
1. Check for UCL sync errors that may be causing Applications to be deleted.
2. Check for operator reconciliation errors that prevent new Applications from being created.
3. Investigate if customer offboarding accounts for the drop.

**Critical** — any runtime loses > 20 % of its Applications in a single sync cycle:
1. Declare incident.
2. Immediately pause the CRA sync loop if possible (scale CRA to 0 replicas) to prevent further deletions.
3. Inspect what the UCL Director returned in the last sync.
4. Restore missing Application CRs from the last known good state or force a full UCL re-sync.

---

### 4.2 Request Throughput Capacity

**Definition:** Maximum sustainable requests per second across Application Gateway (outbound) and Application Connectivity Validator (inbound) under a fixed load-test profile while meeting the latency and error-rate targets defined in sections 3.1 and 3.2.

**How it is measured:**
- Source: Load-test results run against a reference cluster configuration on each release (e.g. using k6 or a similar tool at a fixed concurrency level).
- Production Prometheus rate metrics are collected as a traffic signal only — they are not used for lower-bound health alerts.
- Reported as: sustainable RPS at the defined load profile per release, tracked over time for regression.

**Why it was chosen:**
Throughput capacity establishes the performance envelope of the module under controlled conditions. Using live production traffic for lower-bound alerts would cause false positives during legitimately low-traffic periods. A load-test-derived capacity metric instead catches genuine regressions (e.g. a new version that degrades under load) without false alerts during quiet periods. An unexpected drop in capacity between releases signals a performance regression requiring investigation before broader rollout.

**Thresholds and action plan:**

**Warning** — capacity drops by > 20 % vs. the previous release under the same load profile:
1. Profile the regression: is the bottleneck CPU, memory, or network?
2. Check if recent code changes to CAG or CACV affected the hot path.
3. Investigate HPA settings — confirm they are not limiting scale-out.

**Critical** — capacity drops by > 40 % vs. the previous release, or load-test fails to complete:
1. Block the release.
2. Identify the commit(s) responsible for the regression.
3. Fix and re-run the load test before re-staging the release.

---

## 5. Product Health KPIs

### 5.1 Operator Reconciliation Success Rate

**Definition:** Percentage of ApplicationConnector CR reconciliation cycles that complete without error, as recorded in the CR status.

**How it is measured:**
- Source: Monotonic operator counters for completed reconciliation cycles, labelled by outcome (`success` or `error`); calculate the rate from counter increases over the measurement window. These counters must be exported as Prometheus metrics by the operator.
- Note: the ApplicationConnector CR `.status.state` field alone cannot be used — it is overwritten on each reconcile and exposes neither the number nor the outcome of individual cycles. Polling it would miscount a persistent error state as repeated failures.
- Measured over a rolling 24-hour window.

**Why it was chosen:**
The ACM Kubernetes operator is responsible for the full lifecycle of all module components. A failed reconciliation means the cluster may be running a misconfigured or partially deployed module, which can cause subtle, hard-to-diagnose issues for customers. Since the operator is the control plane of ACM, its health directly determines the reliability of everything else.

**Thresholds and action plan:**

**Warning** — success rate < 99 % over 24 h:
1. Inspect ApplicationConnector CR status conditions and events.
2. Check operator logs for the reconcile error.
3. Common causes: missing Istio CRDs (dependency check), apply conflicts, transient API server errors.
4. Verify all dependency components (Istio) are healthy.

**Critical** — success rate < 95 % over 24 h, or a CR is stuck in `Error` state for > 30 min:
1. Declare incident.
2. Collect operator logs and CR status snapshot.
3. If caused by a bad deployment: roll back.
4. If caused by a broken CR state: delete and recreate the ApplicationConnector CR (last resort, document impact first).
5. Root-cause analysis within 48 business hours.

---

### 5.2 Module Installation / Upgrade Success Rate

**Definition:** Percentage of ACM module installations and upgrades (triggered via Kyma dashboard or lifecycle manager) that complete successfully — reaching `Ready` state — within 15 minutes.

**How it is measured:**
- Source: Lifecycle manager metrics or CI/CD pipeline results (e2e test pass/fail per release).
- For production: ApplicationConnector CR status after a module enable/upgrade operation.
- Tracked per release version.

**Why it was chosen:**
Installation and upgrade success is the first experience a customer has with ACM. A failed installation blocks any value delivery and creates immediate support burden. Tracking this across releases enables early detection of regressions introduced in the packaging, operator startup, or CRD migration path before they reach a large number of customers.

**Thresholds and action plan:**

**Warning** — < 99 % success rate across installations of a given release:
1. Analyse which step fails (dependency check, apply, verify).
2. Check if a specific cluster configuration (Kubernetes version, Istio version) is causing failures.
3. Consider patching the release before broader rollout.

**Critical** — < 95 % success rate, or installation/upgrade failure confirmed on a production customer cluster:
1. Halt further rollout of the affected release.
2. Provide a hotfix or rollback path.
3. Post-mortem required.

---

### 5.3 Time to First Successful Integration

**Definition:** The elapsed time from when a new external system is registered in UCL until the first successful inbound event or outbound API call is confirmed — measured end-to-end.

**How it is measured:**
- Source: Measured in CI/CD integration tests that simulate a new UCL registration and track the timestamp of the first successful event delivery through ACM.
- Reported as the median and p95 over all test runs in a release cycle.
- Target: ≤ 10 minutes (driven by the Runtime Agent's sync interval).

**Why it was chosen:**
This KPI directly reflects the customer experience for the most common integration use case (automatic UCL-based onboarding). A long time-to-first-integration creates friction for customers onboarding new systems and undermines the "easy integration" value proposition of the module. It also catches regressions in the full integration path that unit tests cannot detect.

**Thresholds and action plan:**

**Warning** — p95 > 10 min in CI tests:
1. Profile the integration flow: is the delay in UCL sync, Application CR creation, or event delivery?
2. Check if Runtime Agent sync interval has been inadvertently increased.
3. Check Eventing module latency if the delay is in event delivery.

**Critical** — p95 > 20 min, or the integration test fails entirely:
1. Block the release.
2. Investigate each step of the integration pipeline.
3. If the issue is in ACM: fix and re-run.
4. If the issue is in UCL or Eventing: escalate to the owning team with reproduction steps.

---

## 6. Code Quality & Maintenance KPIs

### 6.1 Test Coverage

**Definition:** Percentage of ACM operator Go code (excluding generated files and test files themselves) covered by unit and integration tests.

**How it is measured:**
- Source: `go test ./... -coverprofile cover.out` output, run in CI on every pull request.
- Coverage is reported per package and as an aggregate.
- Current state (from rewrite investigation): heavy test-to-code ratio (~4.5:1) but with significant gaps, especially in error recovery and deletion paths.

**Why it was chosen:**
The rewrite investigation identified critical test gaps (deletion not verified, error recovery untested, dependency check not covered). Low coverage in these areas means regressions ship silently. This KPI tracks the team's progress in closing those gaps and maintains the discipline needed to keep the codebase maintainable long-term.

**Thresholds and action plan:**

**Warning** — aggregate coverage drops below 80 %, or coverage for `pkg/reconciler` drops below 75 %:
1. Review recent PRs that reduced coverage.
2. Add missing tests before merging the next feature.
3. Prioritise coverage of error paths, specifically: error recovery (deploy fails / patch fails), deletion / finalizer cleanup, `sFnCheckDependencies`, the update path after CR is already Ready, and Compass runtime agent scale-down when its secret is deleted.

**Critical** — aggregate coverage drops below 70 %:
1. Freeze new feature work until coverage is restored.
2. Identify and resolve the largest uncovered blocks.
3. Treat as a quality incident — add to the team's sprint as explicit tasks.

---

### 6.2 Critical Open Bug Age

**Definition:** Number of days a bug labelled `priority/critical` or `priority/high` remains open without a fix merged.

**How it is measured:**
- Source: GitHub Issues with the relevant priority labels in this repository.
- Measured as: age in days from issue creation to PR merge or issue close.
- Reported as median and maximum open age on the weekly team review.

**Why it was chosen:**
Critical bugs left open degrade customer trust and operational stability. Without an explicit KPI, high-priority bugs can accumulate behind feature work. This KPI enforces accountability and ensures that the team maintains a healthy bug backlog.

**Thresholds and action plan:**

**Warning** — any critical/high bug open for > 5 business days:
1. Assign the bug to a specific engineer at the next standup.
2. Break the fix into a minimally scoped PR to avoid stalling.

**Critical** — any critical bug open for > 10 business days:
1. Escalate to team lead.
2. Re-evaluate whether the bug's priority label is accurate — if it truly is critical, it displaces sprint work.
3. Document the blocker explicitly if external dependency.

---

### 6.3 Mean Time to Recovery (MTTR)

**Definition:** Average time from the point an incident is declared (or a critical alert fires) to the point the affected KPI returns within its normal threshold.

**How it is measured:**
- Source: Incident records (GitHub Issues tagged `incident` or equivalent tracking system).
- Measured from the timestamp of the first alert or incident declaration to the timestamp of resolution confirmed.
- Reported as a rolling 90-day average.

**Why it was chosen:**
MTTR is the most direct measure of the team's ability to respond to production issues. A high MTTR amplifies the customer impact of every incident. It also drives investment decisions: if MTTR is high, the team may need better runbooks, on-call tooling, or more automated recovery mechanisms.

**Thresholds and action plan:**

**Warning** — 90-day average MTTR > 4 h:
1. Review the last 3 incidents: which step took the most time?
2. If diagnosis was slow: invest in better dashboards, alerting context, or runbooks.
3. If remediation was slow: identify which manual steps could be automated or pre-approved.

**Critical** — 90-day average MTTR > 8 h, or a single incident with MTTR > 24 h:
1. Hold a dedicated retrospective.
2. Identify systemic gaps: missing runbook, insufficient alerting, unclear on-call responsibility, or external dependency.
3. Create follow-up tasks with owners and due dates.

---

## 7. Measurement Infrastructure

### Data Sources

| Source | Provides |
|--------|---------|
| Prometheus + custom ACM metrics | Gateway latency, error rates, throughput, reconcile counters |
| Kubernetes API (CR status fields) | CompassConnection sync status, ApplicationConnector CR state |
| CI/CD pipeline results | Install/upgrade success rate, test coverage, time-to-first-integration |
| GitHub Issues | Bug age, incident count |
| Incident tracking | MTTR |

### Dashboard

A shared Grafana dashboard should be maintained with:

- One panel per KPI from the overview table (Section 1).
- Traffic-light colour coding matching the warning/critical thresholds.
- A 30-day historical trend view for reliability KPIs.
- An incident log panel showing open incidents and their age.

### Alerting

Prometheus alerting rules should be defined for all KPIs that can be measured in real time (sections 2–5). Alert severity should map to the warning/critical levels defined per KPI. Alerts should fire to the team's primary notification channel and page the on-call engineer for critical alerts.

---

## 8. Review Process and Stakeholder Alignment

### Monthly KPI Review

The team conducts a monthly KPI review meeting covering:

1. **Status review** — traffic-light status for all KPIs against their targets.
2. **Trend analysis** — which KPIs are improving, stable, or degrading over time?
3. **Action item follow-up** — were action items from the previous month completed?
4. **Stakeholder session** — Varban and Artur are invited to the monthly review to:
   - Validate that the KPIs still reflect business objectives.
   - Raise new requirements or adjust thresholds based on customer feedback.
   - Review open incidents and their resolution quality.

### Feedback Loop

This document is a living specification. Changes to KPI definitions, thresholds, or action plans should be:

- Proposed as a pull request to this file.
- Reviewed by at least one team member and one stakeholder (Varban or Artur) before merging.
- Announced to the team in the next standup after merging.

### Continuous Improvement Commitment

The KPIs defined here support continuous improvement by:

- **Proactive issue identification:** Real-time alerting on warning thresholds catches degradation before it becomes critical.
- **Transparency across teams:** The shared dashboard and monthly review ensure that reliability and quality are visible to all stakeholders, not just engineers.
- **Learning from incidents:** MTTR tracking and post-mortems feed directly into improvements to runbooks, alerting, and architecture.
- **Shared understanding:** KPI definitions, measurement methods, and action plans are documented here so any team member can interpret a metric consistently and act without ambiguity.

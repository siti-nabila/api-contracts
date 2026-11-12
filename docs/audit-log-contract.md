# Audit log read contract

`/audit.v1.AuditService/ListAuditLogs` requires TLS, exact-method/IAM-allowed verified service, current ACTIVE unrestricted user/session and exact `iam.audit.read`. Actor filters select historical records; they never identify the caller. User-read/session-read permission and IAM app-access grants are not prerequisites. Audit records do not grant session administration.

Query is minimal: no search, no field selection and no custom sort. PageQuery uses only page (default1), limit (default20/max100), last_id (opaque cursor). Nonempty fields/sort or present search, nil whole request, malformed input/cursor return INVALID_ARGUMENT; nil query/filter uses defaults. Order is `occurred_at DESC, id DESC`. Cursor maximum4096 bytes binds version/full method/fixed ordering/normalized filters and timestamp/BIGINT ID; numeric legacy cursors are invalid. Cursor replaces offset; without cursor page>1 uses offset. Page/limit may change with the same cursor. Total counts pre-cursor matches, total_pages is ceil(total/limit), empty total/pages0, has_next uses an extra row, has_prev follows cursor/page>1, terminal next_cursor is empty. It is one primary snapshot per call, not a frozen export across requests. Metadata overflow returns safe UNAVAILABLE.

All present filters combine with AND. Strings trim whitespace, empty means absent, matching is exact and case-sensitive; unknown values match nothing. actor_user_id present must be1..2147483647. actor_service/action/target_id max100, target_type max50; invalid UTF-8/control characters are invalid. target_id requires target_type; type alone is valid. Both actor selectors are AND when present. No live actor/target lookup is required. Timestamps must be valid; UTC range is [occurred_from, occurred_before), either bound optional, from<before when both present, equivalent at database precision. No default rolling window.

Audit ID is positive BIGINT at most9223372036854775807. Nullable actors/metadata and absent versus present empty snapshots retain presence. Safe snapshots use the typed allowlist below; no raw JSON/map or credential/hash/password/token/header. Malformed stored snapshot, nonobject/null JSON, unknown keys or invalid allowed fields fail the whole page with safe UNAVAILABLE. Session identifiers are descriptive history only.

Invalid service/user/session credentials return UNAUTHENTICATED; wrong method/namespace, restricted actor or missing exact permission returns PERMISSION_DENIED. Dependency/finalization failures return UNAVAILABLE without partial page; cancellation/deadline retain CANCELED/DEADLINE_EXCEEDED, panic is redacted INTERNAL. Existing localized common/auth catalogs are reused without global remapping. Missing data is successful empty only after authorization.

Protected-read admission shares120/minute burst20, cap10000, idle15minutes/cleanup1minute, bounded RetryInfo and RESOURCE_EXHAUSTED/RATE_LIMITED. Invalid requests after verified identity count; unverified identity allocates no user bucket. Successful pages make no audit-success/business/version/session/outbox/idempotency writes. Error attempts retain at most the existing safe independent service audit after read cleanup; recorder failure/panic preserves the original error. Rate rejection makes no DB attempt. Retry performs a fresh authorized read, never persistent replay.

All prior36 methods/eight services and descriptors remain unchanged; adding AuditService makes37 methods/nine services. TesRPC, Register PENDING/token-empty, credential-only Login/Refresh, EdDSA-only and existing session/cursor contracts remain. Deployment/retention/consumer readiness is Phase6 DEP-01, separate from backend completion.

| Message | Field | Tag | Type |
| --- | --- | --- | --- |
| AuditLogFilter | actor_user_id | 1 | optional uint64 |
| AuditLogFilter | actor_service | 2 | string |
| AuditLogFilter | action | 3 | string |
| AuditLogFilter | target_type | 4 | string |
| AuditLogFilter | target_id | 5 | string |
| AuditLogFilter | occurred_from | 6 | google.protobuf.Timestamp |
| AuditLogFilter | occurred_before | 7 | google.protobuf.Timestamp |
| AuditSnapshot | version | 1 | int64 |
| AuditSnapshot | status | 2 | string |
| AuditSnapshot | name | 3 | optional string |
| AuditSnapshot | description | 4 | optional string |
| AuditSnapshot | address | 5 | optional string |
| AuditSnapshot | phone | 6 | optional string |
| AuditSnapshot | role_ids | 7 | repeated uint64 |
| AuditSnapshot | permission_codes | 8 | repeated string |
| AuditSnapshot | application_id | 9 | uint64 |
| AuditSnapshot | session_id | 10 | string |
| AuditSnapshot | reason_code | 11 | string |
| AuditSnapshot | no_op | 12 | bool |
| AuditSnapshot | country_code | 13 | optional string |
| AuditLogItem | id | 1 | uint64 |
| AuditLogItem | actor_user_id | 2 | optional uint64 |
| AuditLogItem | actor_service | 3 | optional string |
| AuditLogItem | action | 4 | string |
| AuditLogItem | target_type | 5 | string |
| AuditLogItem | target_id | 6 | string |
| AuditLogItem | safe_before | 7 | AuditSnapshot |
| AuditLogItem | safe_after | 8 | AuditSnapshot |
| AuditLogItem | reason | 9 | optional string |
| AuditLogItem | request_id | 10 | optional string |
| AuditLogItem | trace_id | 11 | optional string |
| AuditLogItem | occurred_at | 12 | google.protobuf.Timestamp |
| ListAuditLogsRequest | query | 1 | paginator.v1.PageQuery |
| ListAuditLogsRequest | filter | 2 | AuditLogFilter |
| ListAuditLogsResponse | items | 1 | repeated AuditLogItem |
| ListAuditLogsResponse | meta | 2 | paginator.v1.PageMeta |
| ListAuditLogsResponse | next_cursor | 3 | string |

# Application access contract D

Three additive methods extend the thirty-method baseline to exactly33 unary RPCs.
All require TLS, verified IAM/method-allowed service and ACTIVE unrestricted user/session.

| Full method | Exact current permissions |
| --- | --- |
| /user.v1.UserService/ListUserApplicationAccess | iam.user.read |
| /application.v1.ApplicationService/GrantApplicationAccess | iam.user.application.grant AND iam.user.role.grant |
| /application.v1.ApplicationService/RevokeApplicationAccess | iam.user.application.revoke AND iam.user.role.revoke |

## Public historical item

Stored access only, including ACTIVE/REVOKED and disabled/system application history.
Missing pairs are absent; this list does not assert effective permission.

| Field | Type | Presence / sort |
| --- | --- | --- |
| application_id | uint64 | Always included; default ASC and appended tie-breaker |
| application_code | string | ASC/DESC; searchable |
| name | string | ASC/DESC; searchable |
| application_status | string | ACTIVE/DISABLED; ASC/DESC |
| is_system | bool | ASC/DESC |
| status | string | Access ACTIVE/REVOKED; ASC/DESC |
| granted_by | optional uint64 | Nullable; ASC/DESC nulls last |
| granted_at | Timestamp | Known when stored; ASC/DESC |
| revoked_by | optional uint64 | Nullable; ASC/DESC nulls last |
| revoked_at | Timestamp | Nullable; ASC/DESC nulls last |
| revoke_reason | optional string | Nullable; ASC/DESC nulls last |

List accepts PageQuery/PageMeta, default page1/limit20, max100. Application filter,
OR access statuses, literal CONTAINS/PREFIX search over application_code/name
(default both), projection and scalar multi-sort combine with AND. Keywords trim
and max200; percent/underscore are literal. Opaque typed cursors bind target/method/
filters/search/sort, numeric last_id only with application_id-only sort. No cursor
with page>1 uses offset. Total is unique access pairs before cursor/limit; old
catalog and user cursors remain unchanged. Envelope includes user_id/current version.

## Compound mutations

Exactly one idempotency-key metadata value <=200 characters and present positive
int64-compatible expected_user_version are required. Target IDs fit PostgreSQL
INTEGER. Normalize application_code trim/lowercase, length<=50 and registered
syntax; reason trim3-500. Grant accepts1-20 submitted role IDs, canonicalized as a
set; target, business application and every selected role must be ACTIVE and each
role belongs to that application. Grant activates access plus selected roles
atomically and preserves unselected ACTIVE roles; regrant never restores omitted
REVOKED roles. Revoke closes access and every ACTIVE assignment of that application,
including more20 and inactive targets/catalog state. Absent access is not inserted;
remaining ACTIVE roles still revoke. IAM/is_system application mutation is forbidden.

Responses contain safe user/application IDs, optional access_status, actual sorted
changed_role_ids, version and changed. Access-only change can have no changed roles;
role cleanup can have absent access_status. Fresh stale no-op conflicts. Effective
compound command increments user version once, leaves auth_version/sessions/refresh
and other applications unchanged, and stores one application event, audit and safe
result atomically. Fresh full no-op has one NO_OP audit/no event/version increment.
Replay retains original safe result24h after current AND authority and before target
version, without repeated effects. Scope actor/full method/user/application ID/key;
different target independent, conflicting payload ABORTED, expiry atomic new request.
No passwords, hashes or token responses in management/replay/audit/events.

Read quota120/min burst20 and mutation30/min burst5 are shared with baseline methods.
EdDSA/session foundation, PENDING/token-empty Register, generic Login denial, TesRPC,
legacy tags/types and pagination stay compatible. Only Login/Refresh issue credentials.
Live schema/caller/consumer/deployment readiness remains separate Phase6 DEP-01.

Grant/revoke of IAM or system applications returns `APPLICATION_ACCESS_SYSTEM_PROTECTED` (`application.application_access_system_protected`, `AASP`, gRPC `FAILED_PRECONDITION`, HTTP 409). English: “System application access cannot be modified.” Indonesian: “Akses aplikasi sistem tidak dapat diubah.” Existing `SYSTEM_APPLICATION_PROTECTED` and its role-management messages remain unchanged.

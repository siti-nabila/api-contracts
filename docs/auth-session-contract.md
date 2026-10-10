# Session authorization contract prerequisite

This contract is consumed by grpc-auth's auth-session-authorization-foundation
and its separate real Login/EdDSA producer increment. REST/browser changes belong
to rest-orc/frontend. No new RPC is registered by this update.

## Additive fields

| Message | Existing fields preserved | Additions |
| --- | --- | --- |
| UserTokenResponse | token string field 1 | registration_id 2, registration_status 3, refresh_token 4, session_id 5, access_expires_at 6, session_expires_at 7, must_change_password 8 |
| UserData | Fields 1–8, including numeric role_codes | status 9, version 10, roles 11, application_access 12, must_change_password 13 |
| UserListItem | Fields 1–5; int32 ID | status 6, version 7, roles 8, application_access 9 |
| UserFilter | Created range, numeric role_codes, statuses 1–4 | application_code 5 |
| ProfileRequest | name/address/phone 1–3 | field_mask 4 |
| ProfileResponse | profile 1 | version 2, changed 3 |

UserRoleAssignment contains role_id, string RBAC code, display name, assignment
status and role_status. UserApplicationAccess contains application_id, code,
display name, access status and application_status. These safe output messages
carry no password, token, token hash or signing key. Deprecated legacy User
messages are unchanged and are not suitable for new safe list/detail outputs.

Statuses remain strings, consistent with the existing UserFilter: user
PENDING/ACTIVE/SUSPENDED/DEACTIVATED; assignment/access ACTIVE/REVOKED;
role/application ACTIVE/DISABLED. IDs retain their legacy wire types; producers
validate positive IDs fitting PostgreSQL INTEGER, UUID sessions and positive
entity versions. Entity version is distinct from auth_version, which is not
exposed in user/profile list/detail outputs.

Register returns token empty, registration_id and registration_status=PENDING;
refresh/session/expiry fields are absent/empty. It creates no session and grants
no first-user admin privilege. Login leaves registration fields empty and
returns token as access JWT plus refresh_token/session_id/expiry/restricted
fields only after its persistent session/refresh/audit transaction commits.
Defining these fields is not evidence that the producer exists.

Access is at most 15 minutes, capped by absolute session expiry; session/refresh
is at most 7 days since Login with no sliding. JWT verification uses EdDSA,
exact issuer, required sub/sid/jti/iat/exp/auth_version and 30-second JWT skew;
audience is intentionally ignored. Restricted Me emits minimum identity/profile
without roles/application_access. Identity and effective permission come from
authoritative primary state, not these display fields or a browser payload.

For ProfileRequest, absent field_mask preserves full replacement of all three
scalars. A present mask selects only name/address/phone; omitted fields stay
unchanged, selected empty values follow field validation. A present empty mask
is distinct from absence and must not silently invoke full replacement. Unknown
paths are rejected by the consumer. ProfileResponse.version is the committed
user entity version; changed=false represents a validated no-op. These are
service rules, not validation automatically supplied by protobuf.

ListUsers retains auth_id/last_id, sort direction, next_cursor and all existing
page metadata. application_code is an additive application filter. No new
pagination or search semantics are introduced by this contract change.

## Error and retry contract

| Catalog key / Go symbol | Application code | HTTP | gRPC |
| --- | --- | --- | --- |
| auth.service_authentication_failed / ErrServiceAuthenticationFailed | SF | 401 | UNAUTHENTICATED |
| auth.invalid_session / ErrInvalidSession | SI | 401 | UNAUTHENTICATED |
| auth.login_failed / ErrLoginFailed | LF | 401 | UNAUTHENTICATED |
| common.rate_limited / ErrRateLimited | RL | 429 | RESOURCE_EXHAUSTED |

Use one generic login_failed for missing account, wrong password and inactive
account; do not reveal which check failed. Legacy auth.password_mismatch stays
INVALID_ARGUMENT. Exact common unauthorized/forbidden/dependency errors remain
reusable; legacy common.service_unavailable stays UNIMPLEMENTED. Register
CommonCodeMappings and AuthCodeMappings; all existing mappings remain unchanged.
English and Indonesian catalogs are supplied; other locales use existing
fallback behavior. Wrapped private cause text must not enter wire messages.

Limiter consumers first encode common.ErrRateLimited through the catalog, then
attach the standard google.rpc.RetryInfo detail with their computed retry_delay.
Its delay is guidance, not permission to retry credential rotation. If needed,
the existing ErrorInfo metadata may add limit_reason with bounded quota/capacity
values, never account/IP/token/secret data. The existing encoder preserves
detailed statuses on repeated encoding; Decode exposes RetryInfo (nil when
absent) alongside catalog reason/code/http_status. This repository defines no
limiter algorithm, quota or runtime clock.

## Verification and publishing order

1. Generate with make proto. Review the generated diff and keep old fields/RPCs.
2. Run contract wire/descriptor, catalog/locale/mapping/retry tests, all tests and
   vet, also standalone with GOWORK=off. Test grpc-auth and rest-orc in the parent
   workspace to detect source compatibility regressions.
3. Commit/push the tested api-contracts revision only with explicit permission;
   publish an accessible tag or Go pseudo-version from that revision. A working
   tree or workspace build alone is not a published dependency.
4. Explicitly update grpc-auth's go.mod/go.sum to that version, then run tests,
   vet and build both in the workspace and with GOWORK=off against the published
   module. Record actual commit/version, commands, timestamp and results.
5. Only after these steps can EXT-01/task 1.4 be PASS. Real Login/issuer/limiter,
   caller TLS/client readiness and protected cutover remain separate gates.

Do not publish credentials or run migration, deployment, or server startup as a
side effect of contract verification.

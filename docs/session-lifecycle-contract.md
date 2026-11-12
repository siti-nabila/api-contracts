# Session lifecycle contract

Consumer/owner: grpc-auth change `auth-session-lifecycle`. These contracts add
AuthService (RefreshSession, Logout, ChangePassword) and SessionService
(ListMySessions, RevokeMySession, RevokeAllMySessions, ListUserSessions,
RevokeUserSession, RevokeAllUserSessions). Existing user/profile RPC names and field numbers/types
are unchanged; UserListItem.created_at is added at tag10. Contract availability does not mean the backend RPC is deployed.

All methods require TLS and a verified method-allowed `x-service-token`.
RefreshSession proves identity through refresh_token; access JWT is not required.
Refresh returns existing UserTokenResponse credential fields token=1,
refresh_token=4, session_id=5, access_expires_at=6, session_expires_at=7 and
must_change_password=8. registration_id/status stay empty. Access max15 minutes,
session absolute7 days from Login, skew30 seconds does not extend session expiry;
issuer ApplicationName, audience ignored, EdDSA-only. No grace or credential replay.

LogoutRequest.refresh_token is optional: absent selects access-mode authorization
Bearer metadata; present and nonempty selects refresh mode with authorization
absent. Both modes, no proof, duplicate metadata, or present-empty refresh are
invalid. This is a service validation rule, not automatic protobuf validation.
Caller prefers valid access; after access expiry it can use the latest unconsumed
refresh. Consumed refresh is rejected by Logout without triggering RefreshSession
reuse revocation. Latest unconsumed but revoked refresh, or otherwise valid access
proof, can establish terminal ownership until expiry for changed=false Logout only.
Binding/ownership/auth_version still must be valid; this never authorizes another
protected operation. The server validates the selected mode without requiring an
unseen access credential.

ChangePassword uses current_password/new_password, each6–50 characters, verified
self access/session and current password proof. Restricted users may Logout or
ChangePassword; RefreshSession and every SessionService method deny restricted
users. Successful password change revokes all sessions and requires Login again.
No password/management response carries replacement credentials.

Self session methods derive identity from the principal. Administrative lists
require exact iam.session.read; administrative revokes require iam.session.revoke.
Every user_id must be positive and fit PostgreSQL INTEGER; session IDs are
canonical UUIDs bound to the target user. RevokeAllUserSessions additionally
requires present positive expected_user_version (entity version, not auth_version).
Admin revoke metadata idempotency-key is nonempty/max200 characters, with24-hour
safe result retention; current authority precedes replay, replay precedes target
version comparison. Self-service/refresh do not cache credential responses.

Session lists reuse PageQuery/PageMeta, default page1/limit20/max100. Default sort
is created_at DESC, id DESC. Primary sort supports created_at, last_seen_at or
expires_at ASC/DESC, with same-direction session ID tie-breaker and nulls last.
Composite next_cursor binds primary value/null and ID to owner and normalized query;
malformed/query-mismatched cursors are invalid. Unknown query
fields/sorts/search modes are validated by the consuming service. Empty results
have total/total_pages0. Session DTO is a safe allowlist: identifiers, status,
timestamps, bounded display device/IP metadata and is_current. It contains no
credential, hash or auth_version. Backend must propagate context and evaluate
permission/data in one primary snapshot even for empty results.

New errors: auth.refresh_token_reused (RR, HTTP401, UNAUTHENTICATED),
user.version_conflict (VC, HTTP409, ABORTED). Reuse user.idempotency_conflict
(IC, HTTP409, ABORTED), auth.resource_busy, auth.invalid_session,
common.not_found, common.forbidden, common.bad_request and common dependency/time
errors. Current-password mismatch retains auth.password_mismatch/INVALID_ARGUMENT;
legacy mappings are unchanged. Compose CommonCodeMappings, AuthCodeMappings and
UserCodeMappings in the consumer. en/id messages support regional/English fallback.

Refresh reuse requires atomic security mutation/audit/outbox commit before denial.
Consumed-token overlap can cause busy or subsequent reuse revocation; consumers
must single-flight and Login again after lost/uncertain rotation response. Global
revocation/password changes increment user version/auth_version once and do not
reactivate old sessions. This repository declares contracts; runtime checks,
transactions, limiter quotas and persistence belong to grpc-auth.

## ListUsers sorting, filtering and total

UserListItem.created_at is an additive google.protobuf.Timestamp at tag10; tags1..9
retain their names/types. PageQuery.sort and last_id retain their wire contracts.
ListUsers supports id, name, email, address, phone, status, version and created_at
ASC/DESC, plus roles/application_access keyed by the alphabetically smallest name
from each full displayed collection. Empty/null keys are last in either direction.
Ordered multi-sort preserves request order and appends ID ASC when absent; explicit
ID direction is retained. Default ID ASC and auth_id alias remain supported.
Unsupported/internal fields, SQL expressions and duplicate canonical fields are invalid.

Dynamic cursors contain all typed ordered sort values/null markers and ID, bound
to normalized sort/filter/search. Numeric last_id remains accepted only for ID-only
ASC/DESC. Changing sort/filter/search starts with no cursor; an old incompatible
cursor is rejected. Current authority is checked on every request; cursors do not
promise one snapshot across requests or grant access.

Multiple role_codes combine with OR; role category, statuses, application_code,
created range and existing supported keyword search combine with AND. Collection
matching/display uses ACTIVE assignments/access with ACTIVE roles/applications.
Filtering does not trim a matched user's displayed eligible collection. Safe scalar
fields, roles/application_access and created_at are populated according to projection;
projection does not disable query predicates. Total counts unique visible matching
users before cursor/limit, including on an empty later page; join rows or collection
items never increase total. Authority, page, total and collection reads share one
primary snapshot. Runtime validation, registry/query execution and permission checks
remain grpc-auth responsibilities.

Every precommit business/read/conditional/write/cancel/panic error rolls back the
started business transaction. Failure reasons are then recorded safely in a separate
bounded audit transaction; audit failure preserves the original error and emits
bounded telemetry. Proven refresh reuse instead commits security state/audit/outbox
before denial. Insert refresh successor before conditional predecessor update;
conditional failure rolls the successor and all business writes back.

Publishing order: generator, compatibility tests, static analysis and scoped diff
review; authorized commit and normal push to existing feat/user-err, with no tag,
release, merge or force-push; verify module fetch by commit, then update grpc-auth's
build dependency and verify workspace plus GOWORK=off. This is a build reference,
not a runtime pin feature. Local workspace success is not published availability.
Actual rest-orc/deployment readiness remains mandatory Phase6 DEP-01 outside the
backend increment. Contract publication does not complete backend implementation or deployment readiness.

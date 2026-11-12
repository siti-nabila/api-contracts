# Access evaluation contract

The additive /authorization.v1.AuthorizationService/EvaluateAccess unary RPC
extends the existing35-method/seven-service inventory to36 methods/eight services.
Existing descriptors, tags, credential modes and error mappings remain unchanged.

TLS and a verified exact-method service credential are required. The caller must
be allowed for the requested namespace before application/access/permission
lookup. A current ACTIVE unrestricted user/session is required; identity is
verified self, never payload user_id. No administrative permission or IAM
namespace is needed for a permitted business namespace.

| Message | Field | Tag | Type / meaning |
| --- | --- | --- | --- |
| EvaluateAccessRequest | application_code | 1 | string; required trimmed lowercase code, max50 |
| EvaluateAccessRequest | permission_code | 2 | string; required trimmed exact code, max150, matching namespace |
| EvaluateAccessResponse | allowed | 1 | bool; true only for ALLOWED |
| EvaluateAccessResponse | reason | 2 | AccessDecisionReason enum |
| EvaluateAccessResponse | user_id | 3 | uint64; verified current self |
| EvaluateAccessResponse | sid | 4 | string; verified session UUID |
| EvaluateAccessResponse | user_version | 5 | uint64; positive entity version, not a policy revision |
| EvaluateAccessResponse | evaluated_at | 6 | google.protobuf.Timestamp; evaluation instant UTC |

Application syntax is ^[a-z][a-z0-9_]*$. Permission syntax is
^[a-z][a-z0-9_]*([.][a-z][a-z0-9_]*){2,}$; resources may have several segments.
Do not lowercase input automatically. Nil/blank/malformed/wildcard/overlength or
namespace mismatch is INVALID_ARGUMENT. Empty application never means all.

AccessDecisionReason numbers: UNSPECIFIED=0 (never returned), ALLOWED=1,
PERMISSION_MISSING=2, APPLICATION_INACTIVE=3, APPLICATION_ACCESS_REQUIRED=4;
all generated enum names have the ACCESS_DECISION_REASON_ prefix.

After valid identity/namespace: unknown application -> false/PERMISSION_MISSING
(user decision18 October2026); registered disabled application ->
false/APPLICATION_INACTIVE; active business application without ACTIVE user
access -> false/APPLICATION_ACCESS_REQUIRED; otherwise missing/unknown exact
permission -> false/PERMISSION_MISSING; effective permission -> true/ALLOWED.
IAM skips only the user application-access grant gate. Only ACTIVE role and
assignment contributions on ACTIVE applications with nondeleted mappings grant
permission; business access must also be ACTIVE. No numeric-admin/JWT/cache bypass.

All identity/version/gates/contributions use one consistent primary snapshot.
UNAUTHENTICATED identity failures, PERMISSION_DENIED service/restricted failures,
UNAVAILABLE dependency failures, cancellation/deadline and INTERNAL failures are
gRPC errors with no decision response. Use existing common/auth/error/locale
catalogs; enum deny results are not transport errors.

Protected reads share120/minute burst20, existing bounded capacity/TTL/cleanup
and RATE_LIMITED RetryInfo. False decisions attempt one independent safe denied
audit after read cleanup; audit failure never changes the denial. Allowed reads
write no success audit. No domain state/version/auth_version/session/refresh,
outbox or idempotency result changes. Rate denial uses bounded telemetry.

Responses expose no password/hash/access/refresh credentials or complete role
lists. Sid is not a credential. Results are not cacheable/reusable access tickets;
user_version is not a policy revision. Response loss/retry always evaluates anew.
Domain callers still check resource authorization near their protected action.
Deployment/consumer readiness remains Phase6 DEP-01; this contract alone does
not deploy, publish, migrate or invalidate credentials.

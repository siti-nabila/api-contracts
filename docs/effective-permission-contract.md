# Effective permission reads E

Two additive unary methods extend the33-method baseline to35. Both require TLS,
verified exact-method service authentication and an ACTIVE unrestricted actor/session.

| Full method | Identity and authority |
| --- | --- |
| /user.v1.UserService/GetMyEffectivePermissions | Verified self only; no iam.user.read required |
| /user.v1.UserService/GetUserEffectivePermissions | IAM service namespace plus exact current iam.user.read, including self-target |

## Messages and tags

| Message | Field | Tag | Type |
| --- | --- | --- | --- |
| GetMyEffectivePermissionsRequest | application_code | 1 | string |
| GetUserEffectivePermissionsRequest | user_id | 1 | uint64 |
| GetUserEffectivePermissionsRequest | application_code | 2 | string |
| EffectivePermissionRole | role_id | 1 | uint64 |
| EffectivePermissionRole | code | 2 | string |
| EffectivePermission | code | 1 | string |
| EffectivePermission | roles | 2 | repeated EffectivePermissionRole |
| GetEffectivePermissionsResponse | user_id | 1 | uint64 |
| GetEffectivePermissionsResponse | user_status | 2 | string |
| GetEffectivePermissionsResponse | version | 3 | uint64 |
| GetEffectivePermissionsResponse | permissions | 4 | repeated EffectivePermission |

Blank/trimmed-empty application_code returns only namespaces allowed for the
verified service caller. Nonempty code must match lowercase ^[a-z][a-z0-9_]*$
with length<=50; malformed input returns INVALID_ARGUMENT. Explicit disallowed
namespace returns PERMISSION_DENIED before existence lookup. Allowed unknown or
DISABLED application yields an empty collection for an otherwise authorized
existing target. User IDs must fit1..2147483647; missing/soft-deleted targets return
NOT_FOUND after actor authorization. A self request has no target user_id.

Actor authority, target status/version and contributions use one primary snapshot.
Union includes only ACTIVE assignments/roles/applications, same-application mappings,
and ACTIVE business application access; IAM needs no application grant. Each code
is unique and ascending, each contributing role ID/code pair unique and ID ascending.
No pagination, truncation, search, wildcard or numeric-role/JWT/cache bypass.

For nondeleted PENDING/SUSPENDED/DEACTIVATED targets, admin receives still-ACTIVE-role
contributions under the same application gates, with required user_status. This is
descriptive retained configuration, not audit reconstruction or a usable access
grant. REVOKED/DISABLED sources remain excluded. No target session is required or
reactivated. Version is informational entity version, not a permission cache key.

Responses expose no password/hash/token/refresh/session identifier. Reads mutate no
domain/session/version/audit-success/outbox/idempotency result. Retry after response
loss performs a fresh authorized read. Failures retain safe independent attempt
audit after cleanup; dependency failure is UNAVAILABLE, cancellation/deadline retain
CANCELED/DEADLINE_EXCEEDED. Existing auth/common error catalogs/locales are reused.

Both methods share protected-read quota120/min burst20 and existing RetryInfo and
bounded state policy. Existing33 methods/tags/types/credential modes, EdDSA/session
rules, restricted Me/ChangePassword/Logout allowlist and pagination remain unchanged.
Actual deployment/consumer readiness remains Phase6 DEP-01.

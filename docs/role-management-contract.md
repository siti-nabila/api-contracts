# Role management and catalog contract B

These additive contracts target grpc-auth increment B. Contract availability
does not imply backend registration, live schema readiness or deployment PASS.
The producer implements the rules below; protobuf serialization alone does not
validate business input. Publish the verified module before updating grpc-auth,
then verify the consumer both in its workspace and with `GOWORK=off`.

## Methods and authority

Every method requires TLS and a verified `x-service-token` allowed for its exact
full method and IAM namespace, plus an ACTIVE unrestricted user/access/session.
Caller identity comes from verified context, never a payload target ID. Resolve
current exact permission in the primary read snapshot or after ordered mutation
locks; authorization precedes existence disclosure and idempotency replay.

| Full method | Exact permission | Admission | Key metadata |
| --- | --- | --- | --- |
| /role.v1.RoleService/ListRoles | iam.role.read | protected-read | None |
| /role.v1.RoleService/GetRole | iam.role.read | protected-read | None |
| /role.v1.RoleService/CreateRole | iam.role.create | admin-mutation | Required |
| /role.v1.RoleService/UpdateRole | iam.role.update | admin-mutation | Optional |
| /role.v1.RoleService/ChangeRoleStatus | iam.role.update | admin-mutation | Required |
| /role.v1.RoleService/ReplaceRolePermissions | iam.role.permission.update | admin-mutation | Required |
| /permission.v1.PermissionService/ListPermissions | iam.permission.read | protected-read | None |
| /application.v1.ApplicationService/ListApplications | iam.permission.read | protected-read | None |

The groups share existing quotas: protected-read 120/minute burst20,
admin-mutation 30/minute burst5. Applicable invalid input consumes admission.
Rejection returns bounded RetryInfo/RESOURCE_EXHAUSTED without business writes.
Catalog target application_code does not replace IAM authority. No permission,
application or role deletion/assignment RPC is added in this increment.

## Safe reads and impact

ListRoles exposes RoleSummary; GetRole returns that summary and registered
permissions ordered code ASC/id ASC, including disabled/system roles and disabled
applications. Unassigned roles remain visible. GetRole missing target returns
NOT_FOUND after authority. Permission/application lists expose catalog metadata,
without granting permission or application access. There are no credentials,
hashes, auth_version, or user identities behind impact counts in these messages.

Role description and created_at/updated_at may be absent for SQL NULL; timestamps
are not fabricated for migrated rows. String code and numeric role_code remain
distinct. affected_user_count counts DISTINCT existing user IDs whose assignment
to the role is ACTIVE, including PENDING/ACTIVE/SUSPENDED/DEACTIVATED or soft-deleted
existing accounts, regardless of role/application/access state. REVOKED assignments
do not count; permission joins never multiply the count. List/detail count agree.

All authority, collection, total, count and detail reads use one read-only primary
repeatable-read snapshot. Dependency failures return UNAVAILABLE without partial
results; cancellation/deadline mappings remain intact. Reads change no entity
versions and emit no domain events. Catalog is separate from user filter-options.

## Query, projection and cursor

Lists use paginator.v1.PageQuery and PageMeta, response items field1, meta field2,
next_cursor field3. Default page1/limit20, maximum limit100. Total counts unique
matching roots before cursor/offset/limit. Empty results have total/total_pages0
and empty items/next_cursor. Without cursor page>1 uses deterministic offset;
with cursor page is ordinal metadata and does not add offset. Fetch limit+1 to
derive has_next; has_prev is true with a cursor or page>1.

Role filters: application_code plus repeated ACTIVE/DISABLED statuses. Permission
filter: application_code. Application filter: repeated ACTIVE/DISABLED statuses.
Absent filters include all states. Status alternatives combine OR; categories
and keyword combine AND. A valid nonexistent application filter yields an empty
page. Permission has no status field.

Search trims keyword (at most200 characters), empty means no search. UNSPECIFIED
or CONTAINS performs case-insensitive literal contains; PREFIX performs literal
prefix. Escape percent/underscore/backslash. Selected search fields combine OR.
Unsupported mode/field/status fails INVALID_ARGUMENT. FULL_TEXT/TRIGRAM/combined
modes are not supported by these catalogs.

| Catalog | Default search | Allowed search fields |
| --- | --- | --- |
| Role | code, name | code, name, description, application_code |
| Permission | code, description | code, description, application_code |
| Application | code, name | code, name |

Every public scalar below supports ASC/DESC; ordered multi-sort rejects duplicate
and unknown fields, appends id ASC when absent, and defaults to id ASC. Text uses
deterministic PostgreSQL C collation consistently in ordering and comparisons.
Nullable values are nulls last in both directions. Requests never supply SQL
expressions or internal column names. The producer owns explicit SQL registries.

| Item | Sortable/projection field | Type | Nullable |
| --- | --- | --- | --- |
| RoleSummary | id | uint64 | No |
| RoleSummary | code | string | No |
| RoleSummary | name | string | No |
| RoleSummary | description | string | Yes |
| RoleSummary | application_id | uint64 | No |
| RoleSummary | application_code | string | No |
| RoleSummary | status | string | No |
| RoleSummary | is_system | bool | No |
| RoleSummary | version | uint64 | No |
| RoleSummary | role_code | int32 | No |
| RoleSummary | created_at | google.protobuf.Timestamp | Yes |
| RoleSummary | updated_at | google.protobuf.Timestamp | Yes |
| RoleSummary | affected_user_count | uint64 | No |
| PermissionItem | id | uint64 | No |
| PermissionItem | application_id | uint64 | No |
| PermissionItem | application_code | string | No |
| PermissionItem | code | string | No |
| PermissionItem | description | string | No |
| ApplicationItem | id | uint64 | No |
| ApplicationItem | code | string | No |
| ApplicationItem | name | string | No |
| ApplicationItem | status | string | No |
| ApplicationItem | is_system | bool | No |

PageQuery.fields absent means all; otherwise selected fields plus id are returned.
Unknown/duplicate projection fields fail. Hidden sort values remain available for
cursor construction even if not projected. Future public scalars require an
explicit producer registry entry and tests before exposure.

last_id carries a versioned base64url composite cursor with catalog method,
normalized filter/search/sort digest, ordered typed sort values/null markers and
ID. Digest excludes page/limit/projection. Mixed-direction lexicographic keyset
comparison includes null buckets and ID tie-breaker. Malformed/type/query/method
mismatches fail INVALID_ARGUMENT. Positive numeric last_id is also supported for
ID-only sort. Changed query starts without a cursor. Terminal next_cursor is
empty. Cursor is not authority and does not retain a snapshot between requests.

## Mutations, presence and versions

New IDs are 1..2147483647. Required expected_role_version is present and
1..9223372036854775807; it is role entity version, not user auth_version.
Names/descriptions/codes/reasons are trimmed. Canonical code is lower-case as
submitted, never silently case-folded. Role code is at most100 characters and
matches `^[a-z][a-z0-9_]*[.][a-z][a-z0-9_]*$`; application code matches its first
segment. Permission codes are at most150 characters and use registered catalog
grammar application.resource.action, with additional resource segments allowed.
Permission namespace AND application_id must both match the role application.
There is no twenty-item permission-set cap; the transport request cap still applies.

| Command | Input and resulting behavior |
| --- | --- |
| CreateRole | Existing ACTIVE nonsystem business application; unique code; trimmed name1..50; optional description0..100; registered same-app permission_codes deduped as a final set, empty allowed. Server creates ACTIVE/is_system=false/version1 and allocates positive numeric role_code from DB sequence. No assignments/access/credentials created. |
| UpdateRole | Only optional name/description and matching nonempty field_mask. Omitted fields unchanged; selected empty description clears to SQL NULL. Reject duplicate/unknown/immutable mask paths and mismatched presence. Name remains1..50. Required expected_role_version. |
| ChangeRoleStatus | Explicit optional status presence required, ACTIVE or DISABLED; reason trimmed3..500 and required expected_role_version. |
| ReplaceRolePermissions | Required PermissionSet message: omitted intent invalid, present empty codes clears all mappings. Final same-app registered set deduped; reason trimmed3..500 and required expected_role_version. |

System, IAM and legacy IAM roles are protected even for nominal no-ops. Existing
custom roles remain editable in disabled applications. Disable/replace changes
effective permission in the next authoritative evaluation without changing user
version/auth_version, sessions/refresh, assignments or application access. Enable
does not restore revoked assignments; overlapping eligible roles still contribute.

Current authority precedes safe replay; replay precedes version comparison. New
stale requests return VERSION_CONFLICT before no-op. Effective mutation increments
role version exactly once. Valid new no-op returns changed=false, one NO_OP audit,
no version/event; this includes unkeyed UpdateRole. Replay adds no effects/audit.

## Idempotency and atomic outcomes

idempotency-key is metadata, exactly one nonempty value at most200 characters
when provided. CreateRole/ChangeRoleStatus/ReplaceRolePermissions require it;
UpdateRole accepts it optionally. Scope is verified actor + exact full method +
role-code/{canonical_code} for creation or role/{id} for updates + key. Fingerprint
includes normalized payload/presence/expected version and sorted unique permission
set, excluding correlation/deadline/credential. Store exact safe original response
for24h, including nullable metadata/count. Replay returns it even if target later
changes, after fresh authorization. Same scope/key with different payload returns
IDEMPOTENCY_CONFLICT; other targets are independent. Expiry is handled atomically.

| Command | Required success audit | Pending event |
| --- | --- | --- |
| CreateRole | ROLE_CREATED | iam.role.created.v1 |
| UpdateRole | ROLE_UPDATED | iam.role.updated.v1 |
| ChangeRoleStatus | ROLE_UPDATED | iam.role.updated.v1 |
| ReplaceRolePermissions | ROLE_PERMISSIONS_CHANGED | iam.role.permissions_changed.v1 |

Role/mapping/version/audit/outbox/keyed safe result commit together. No failure
leaves partial writes; independent safe failure auditing follows cleanup and
preserves the original error on audit failure. Commit uncertainty returns a safe
error without automatic retry or claiming certain rollback. Events contain only
allowlisted role/application/status/final permission data, controlled reason code,
committed aggregate version and existing envelope; free-form reason is audit only.
Broker availability is not required; published_at remains NULL pending delivery.

## Error catalog and compatibility

Register mapping.RoleCodeMappings and mapping.ApplicationCodeMappings alongside
existing mappings; register role.Registry/application.Registry for decoding.
Each new key supplies en/id with English fallback and regional base fallback.
Wrapping retains source identity/cause; encoded errors expose safe catalog text.

| Qualified key | Application code | HTTP | gRPC |
| --- | --- | --- | --- |
| role.role_not_found | RNF | 404 | NOT_FOUND |
| role.role_code_already_exists | RCE | 409 | ALREADY_EXISTS |
| role.system_role_protected | SRP | 409 | FAILED_PRECONDITION |
| role.invalid_permission_set | IPS | 400 | INVALID_ARGUMENT |
| role.version_conflict | VC | 409 | ABORTED |
| role.idempotency_conflict | IC | 409 | ABORTED |
| role.resource_busy | RB | 409 | ABORTED |
| application.application_not_found | ANF | 404 | NOT_FOUND |
| application.application_inactive | AI | 409 | FAILED_PRECONDITION |
| application.system_application_protected | SAP | 409 | FAILED_PRECONDITION |

Reuse existing field validation/common authorization/dependency/cancellation/
deadline errors; unknown failures stay generic INTERNAL. Existing auth/user/common
codes and mappings remain unchanged, including legacy ServiceUnavailable mapping.
All19 baseline methods and entire old descriptors remain unchanged; eight unary
methods add up to27. TesRPC, numeric user role_codes and session contracts remain.
The baseline descriptor fixture and tests/protobuf role scenarios verify wire/tag/
type/presence compatibility. New protobuf wrappers preserve absent versus empty
permission intent through binary and JSON. Producer tests separately verify SQL,
authority, snapshots, mutation semantics and live dedicated-test-DB correctness.

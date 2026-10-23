# User management contract

Four additive unary RPCs belong to `user.v1.UserService`: GetUser, CreateUser,
UpdateUser, and ChangeUserStatus. Existing method and field identities remain.
The IAM permissions are respectively `iam.user.read`, `iam.user.create`,
`iam.user.update`, and `iam.user.status.update`. Authentication and business
transactions belong to grpc-auth, not this contract repository.

CreateUser/GetUser retain their existing request/response wrappers. `User` tags
1–6 are unchanged; legacy `password` tag 4 MUST be empty in every management
response. Safe metadata adds address(7), status(8), version(9), created_at(10),
must_change_password(11), roles(12), and application_access(13), and country_code(14). Phone uses legacy
phone_no(6). No auth_version, token, refresh token, or password hash is returned.
Collections show ACTIVE assignments/access with ACTIVE roles/applications.

CreateUser retains name(1), email(2), password(3); adds explicit optional status(4),
address(5), phone(6), and optional country_code(7). Omitted/empty status is invalid; choose PENDING or ACTIVE.
Email is trimmed/lowercased. Email/password lengths are 6–50, supplied name 3–100,
address up to 250, phone up to 20. Creation starts version 1 with
must_change_password=true and issues no credentials. Initial password delivery is
external to this contract; it is never stored in replay results.

UpdateUser takes positive user_id(1), present positive expected_user_version(2),
optional name(3)/address(4)/phone(5), field_mask(6), and optional country_code(7). Only those four profile
paths are accepted. Mask selection and field presence must agree. Omitted fields
are unchanged; selected empty address/phone clears the field. Selected empty name
is invalid. Response user(1)/changed(2) contains the committed safe state.

ChangeUserStatus takes positive user_id(1), present status(2), reason(3), and
present positive expected_user_version(4). Response user_id(1), status(2),
version(3), changed(4). PENDING can become ACTIVE or DEACTIVATED; ACTIVE can become
SUSPENDED or DEACTIVATED; SUSPENDED can become ACTIVE or DEACTIVATED; DEACTIVATED
can become ACTIVE. A same-status request with matching version is a no-op.
Self suspend/deactivate and removing the last active platform admin are denied.

CreateUser/ChangeUserStatus require idempotency metadata; UpdateUser permits it.
Keys are nonempty and at most 200 characters, scoped to actor + exact full method
+ normalized account/user target. Identical retained requests replay the initial
safe response for 24 hours after current authorization; differing requests
conflict. Target version is checked after replay, before no-op detection.
Actor, roles, permissions and application access cannot be set through requests.

Existing auth.not_found (NotFound/404), auth.already_exists (AlreadyExists/409),
auth validation, user.version_conflict, user.resource_busy and
user.idempotency_conflict (Aborted/409) are reused without remapping. New errors:
user.invalid_status_transition (IST, FailedPrecondition/409),
user.self_inactivation_forbidden (SIF, PermissionDenied/403), and
user.last_admin_required (LAR, FailedPrecondition/409). English/Indonesian catalogs
and existing locale fallback apply. No raw internal failure details are exposed.

Publication order: regenerate protobuf, test descriptors/presence/legacy wire
compatibility/error metadata/locales, run full tests and vet, review source and
generated diffs, then publish a reviewed commit before updating grpc-auth. No
runtime business guarantees are inferred from descriptor tests alone.

See [country-code-contract.md](country-code-contract.md) for nullable calling-code validation, clearing, legacy omission, audit and replay compatibility.

# Me Access Snapshot Contract

`/user.v1.UserService/Me` retains `google.protobuf.Empty` input and `user.v1.UserData` output. TLS/service authentication and the current user's valid session are required; no admin permission, target ID or application filter is added.

`UserData` fields 1–14 retain their tags/types. New field `map<string, bool> permissions = 15` contains only current effective grants (`true`) within namespaces allowed for the verified service caller. Missing keys are not granted. Empty maps/collections may be omitted by protobuf JSON serializers and never imply default access. Map iteration/wire order is unspecified.

Safe fields include identity/profile, account status, user entity version, user creation time, restriction marker, roles and application access. User version is not a permission-policy version. Restricted Me has no legacy/new role, application or permission privileges. No credential, session ID or auth_version is added.

**Intentional Me.phone semantic change:** grpc-auth returns the literal stored `country_code + phone`, with NULL/empty country code treated as empty. No prefix stripping, zero removal, separator or default country is applied. `+62` plus `+62812` returns `+62+62812`; existing data is not rewritten. `country_code` remains separate. Profile/detail/list phone semantics are unchanged. Consumers must not concatenate the Me phone again.

The response is a current read from one primary snapshot, not a reusable authorization token. Route permission is server-owned; browser-provided maps do not authorize backend actions. Existing permission/evaluation RPCs remain available. Publish the additive contract before updating consumer dependency pins; local workspace compatibility does not prove a published version exists.

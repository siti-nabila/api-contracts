# Profile telephone calling code

`country_code` is a nullable telephone calling code such as `+62`, not an ISO
country identifier. It is stored separately from `phone`; adding it does not
rewrite existing phone numbers, split prefixes, or assign a default country.

| Message | Field | New tag | Type |
| --- | --- | --- | --- |
| profile.v1.Profile | country_code | 6 | string |
| profile.v1.ProfileRequest | country_code | 5 | optional string |
| user.v1.User | country_code | 14 | string |
| user.v1.UserData | country_code | 14 | string |
| user.v1.UserListItem | country_code | 11 | string |
| user.v1.CreateUserRequest | country_code | 7 | optional string |
| user.v1.UpdateUserRequest | country_code | 7 | optional string |
| audit.v1.AuditSnapshot | country_code | 13 | optional string |

All previous field numbers, types, service names and RPC signatures stay intact.
Input presence distinguishes omitted values from an explicit empty string.
Empty/omitted creation values map to SQL NULL. Empty selected update values
clear the code to SQL NULL. NULL database values appear as an empty string in
profile/user responses; no `+62` fallback is introduced.

UpdateProfile accepts `country_code` in its field mask and requires the optional
field to be present when selected. Without a field mask, the three legacy
profile fields retain full-replacement behavior, while an omitted country code
is preserved for older callers. UpdateUser requires field presence and mask
selection to agree, including the new field.

grpc-auth accepts empty values or a plus sign followed by one to three ASCII
digits with a nonzero first digit. This is syntax validation, not a registry of
assigned calling codes. Ownership, version checks, transaction boundaries,
audit and idempotency behavior remain in grpc-auth. Country-code changes enter
profile snapshots, safe replay results and request fingerprints. Historical
snapshots/replay records without this field remain readable.

Publication order: regenerate with `make proto`, run tests/vet, publish the
reviewed contract version, then pin that version in grpc-auth and validate with
`GOWORK=off`. Workspace tests alone do not prove published-version compatibility.

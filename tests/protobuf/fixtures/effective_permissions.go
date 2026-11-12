package fixtures

import _ "embed"

// EffectivePermissionsBaseline freezes all 33 methods and complete pre-E messages.
//
//go:embed effective_permissions_baseline.json
var EffectivePermissionsBaseline []byte

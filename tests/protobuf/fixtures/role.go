package fixtures

import _ "embed"

// RoleBaseline captures every existing descriptor before increment B.
//
//go:embed role_baseline.json
var RoleBaseline []byte

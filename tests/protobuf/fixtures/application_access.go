package fixtures

import _ "embed"

// ApplicationAccessBaseline captures all 30 pre-D methods and complete messages.
//
//go:embed application_access_baseline.json
var ApplicationAccessBaseline []byte

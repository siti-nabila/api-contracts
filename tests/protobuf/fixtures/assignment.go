package fixtures

import _ "embed"

// AssignmentBaseline captures all 27 pre-C methods and their source descriptors.
//
//go:embed assignment_baseline.json
var AssignmentBaseline []byte

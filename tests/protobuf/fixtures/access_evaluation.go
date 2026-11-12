package fixtures

import _ "embed"

// AccessEvaluationBaseline freezes the complete pre-evaluation descriptors.
//
//go:embed access_evaluation_baseline.json
var AccessEvaluationBaseline []byte

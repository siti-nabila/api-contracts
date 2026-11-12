package fixtures

import _ "embed"

// AuditLogReadsBaseline freezes the complete 36-method pre-audit contract.
//
//go:embed audit_log_reads_baseline.json
var AuditLogReadsBaseline []byte

package fixtures

const (
	ValidCatalog = `errors:
  missing:
    code: NF
    http_status: 404
    en: Not found.
    id: Tidak ditemukan.
    zh: "未找到。"
`

	UnknownFieldCatalog = `errors:
  missing:
    code: NF
    http_status: 404
    en: Not found.
    invalid@locale: invalid
`

	GRPCCodeCatalog = `errors:
  state:
    code: ST
    http_status: 400
    grpc_code: FAILED_PRECONDITION
    en: Invalid state.
`
)

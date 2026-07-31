package fixtures

const (
	ValidCatalog = `errors:
  missing:
    code: NF
    http_status: 404
    en: Not found.
    id: Tidak ditemukan.
`

	OverrideCatalog = `errors:
  state:
    code: ST
    http_status: 400
    grpc_code: FAILED_PRECONDITION
    en: Invalid state.
    id: Status tidak valid.
`

	UnknownFieldCatalog = `errors:
  missing:
    code: NF
    http_status: 404
    en: Not found.
    extra: invalid
`

	InvalidGRPCCodeCatalog = `errors:
  state:
    code: ST
    http_status: 400
    grpc_code: SOMETHING_NEW
    en: Invalid state.
`
)

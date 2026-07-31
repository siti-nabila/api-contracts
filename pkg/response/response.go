package response

type Success struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type FieldError struct {
	Code   string              `json:"code"`
	Errors map[string][]string `json:"errors"`
	Data   []any               `json:"data"`
}

type Description struct {
	Description string `json:"description"`
}

type DescriptionError struct {
	Code   string      `json:"code"`
	Errors Description `json:"errors"`
	Data   []any       `json:"data"`
}

func EmptyData() []any {
	return make([]any, 0)
}

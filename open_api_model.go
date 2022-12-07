package grest

type OpenAPIModel interface {
	Tags() []string
	Summary() string
	Description() string
	Accept() string
	Produce() string
	Security() []string
	RequestBody() any
	// SuccessResponses() []Response
	// FailureResponses() []Response
	ExternalDoc() (string, string)
}

func (Model) Tags() []string {
	return []string{}
}

func (Model) Summary() string {
	return ""
}

func (Model) Description() string {
	return ""
}

func (Model) Accept() string {
	return "json"
}

func (Model) Produce() string {
	return "json"
}

func (Model) Security() []string {
	return []string{}
}

func (Model) RequestBody() any {
	return nil
}

func (Model) ExternalDoc() (string, string) {
	url := ""
	description := ""
	return url, description
}

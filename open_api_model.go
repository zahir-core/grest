package grest

type OpenAPIModelInterface interface {
	OpenAPITags() []string
	OpenAPISummary() string
	OpenAPIDescription() string
	OpenAPIAccept() string
	OpenAPIProduce() string
	OpenAPIPathParam() []map[string]any
	OpenAPIHeaderParam() []map[string]any
	OpenAPICookieParam() []map[string]any
	OpenAPIQueryParam() []map[string]any
	OpenAPIBody() any
	OpenAPISecurity() []map[string][]string
	OpenAPIExternalDoc() (string, string)
}

func (Model) OpenAPITags() []string {
	return []string{}
}

func (Model) OpenAPISummary() string {
	return ""
}

func (Model) OpenAPIDescription() string {
	return ""
}

func (Model) OpenAPIAccept() string {
	return "json"
}

func (Model) OpenAPIProduce() string {
	return "json"
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
func (Model) OpenAPIPathParam() []map[string]any {
	h := []map[string]any{}
	// example :
	// h = append(h, map[string]any{
	// 	"in":   "path",
	// 	"name": "contactID",
	// 	"schema": map[string]any{
	// 		"type": "string",
	// 	},
	// })
	return h
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
func (Model) OpenAPIHeaderParam() []map[string]any {
	h := []map[string]any{}
	// example :
	// h = append(h, map[string]any{
	// 	"in":   "header",
	// 	"name": "Content-Language",
	// 	"schema": map[string]any{
	// 		"type": "string",
	// 	},
	// 	"examples": map[string]any{
	// 		"English (US)": map[string]any{
	// 			"value": "en-US",
	// 		},
	// 		"Bahasa Indonesia": map[string]any{
	// 			"value": "id-ID",
	// 		},
	// 	},
	// })
	return h
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
func (Model) OpenAPICookieParam() []map[string]any {
	h := []map[string]any{}
	// example :
	// h = append(h, map[string]any{
	// 	"in":   "header",
	// 	"name": "token",
	// 	"schema": map[string]any{
	// 		"type": "string",
	// 	},
	// })
	return h
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
func (Model) OpenAPIQueryParam() []map[string]any {
	q := []map[string]any{}
	q = append(q, map[string]any{
		"in":   "query",
		"name": "params",
		"schema": map[string]any{
			"type": "object",
			"additionalProperties": map[string]any{
				"type": "string",
			},
		},
		"style":   "form",
		"explode": true,
	})
	return q
}

func (Model) OpenAPIBody() any {
	return nil
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#response-object
func (m Model) OpenAPIResponses() map[string]any {
	res := map[string]any{}
	// example :
	// res["200"] = map[string]any{
	// 	"model": m,                      // will auto create related components for content schema $ref: '#/components/schemas/Model'
	// 	"description": "Success",
	// }
	// res["401"] = map[string]any{
	// 	"model": app.UnauthorizedModel,  // will auto create related components for content schema $ref: '#/components/schemas/UnauthorizedModel'
	// 	"description": "Success",
	// }
	return res
}

func (Model) OpenAPISecurity() []map[string][]string {
	sec := []map[string][]string{}
	// example :
	// sec = append(sec, map[string][]string{
	// 	"petstore_auth": {
	// 		"write:pets",
	// 		"read:pets",
	// 	},
	// })

	return sec
}

func (Model) OpenAPIExternalDoc() (string, string) {
	url := ""
	description := ""
	return url, description
}

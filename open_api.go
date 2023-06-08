package grest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// The full Latest OpenAPI Specification is available on https://spec.openapis.org/oas/latest.html
type OpenAPI struct {
	OpenAPI           string              `json:"openapi,omitempty"`
	Info              OpenAPIInfo         `json:"info,omitempty"`
	JsonSchemaDialect string              `json:"jsonSchemaDialect,omitempty"`
	Servers           []map[string]any    `json:"servers,omitempty"`
	Paths             map[string]MapSlice `json:"paths,omitempty"`
	Webhooks          map[string]any      `json:"webhooks,omitempty"`
	Components        map[string]any      `json:"components,omitempty"`
	Security          []map[string]any    `json:"security,omitempty"`
	Tags              []map[string]any    `json:"tags,omitempty"`
	ExternalDocs      OpenAPIExternalDoc  `json:"externalDocs,omitempty"`
}

func (o *OpenAPI) SetVersion() {
	o.OpenAPI = "3.0.3"
}

func (o *OpenAPI) Configure() {
	// add your openapi doc here
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#server-object
func (o *OpenAPI) AddServer(server map[string]any) {
	o.Servers = append(o.Servers, server)
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#tag-object
func (o *OpenAPI) AddTag(tag map[string]any) {
	tagName, ok := tag["name"].(string)
	if ok {
		isTagNameExists := false
		for _, tg := range o.Tags {
			existingTagName, _ := tg["name"].(string)
			if tagName == existingTagName {
				isTagNameExists = true
			}
		}
		if !isTagNameExists {
			o.Tags = append(o.Tags, tag)
		}
	}
}

func (o *OpenAPI) AddPath(key, method string, operationObject any) {
	if o.Paths != nil {
		path, _ := o.Paths[key]
		o.Paths[key] = append(path, map[string]any{"key": method, "value": operationObject})
	} else {
		o.Paths = map[string]MapSlice{key: {map[string]any{"key": method, "value": operationObject}}}
	}
}

func (o *OpenAPI) AddWebhook(key string, val any) {
	if o.Webhooks != nil {
		o.Webhooks[key] = val
	} else {
		o.Webhooks = map[string]any{key: val}
	}
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#components-object
func (o *OpenAPI) AddComponent(key string, val any) {
	if o.Components != nil {
		component, isComponentExists := o.Components[key]
		c, cOk := component.(map[string]any)
		v, vOk := val.(map[string]any)
		if isComponentExists && cOk && vOk {
			for name, value := range v {
				_, isNameExists := c[name]
				if !isNameExists {
					c[name] = value
				}
			}
			o.Components[key] = c
		} else {
			o.Components[key] = val
		}
	} else {
		o.Components = map[string]any{key: val}
	}
}

func (o *OpenAPI) AddRoute(path, method string, op OpenAPIOperationInterface) {
	fmt.Println("OpenAPI : add paths", path, method)
	operationObject := map[string]any{}
	if len(op.OpenAPITags()) > 0 {
		tags := op.OpenAPITags()
		for _, tagName := range tags {
			o.AddTag(map[string]any{"name": tagName})
		}
		operationObject["tags"] = tags
	}
	if op.OpenAPISummary() != "" {
		operationObject["summary"] = op.OpenAPISummary()
	}
	if op.OpenAPIDescription() != "" {
		operationObject["description"] = op.OpenAPIDescription()
	}
	externalDocUrl, externalDocDesc := op.OpenAPIExternalDoc()
	if externalDocUrl != "" || externalDocDesc != "" {
		operationObject["externalDocs"] = map[string]any{
			"url":         externalDocUrl,
			"description": externalDocDesc,
		}
	}
	if op.OpenAPIOperationID() != "" {
		operationObject["operationId"] = op.OpenAPIOperationID()
	}

	params := []map[string]any{}
	params = append(params, op.OpenAPIPathParam()...)
	params = append(params, op.OpenAPIHeaderParam()...)
	params = append(params, op.OpenAPICookieParam()...)
	params = append(params, op.OpenAPIQueryParam()...)
	if len(params) > 0 {
		operationObject["parameters"] = params
	}

	requestBody := op.OpenAPIBody()
	if requestBody != nil {
		for k, m := range requestBody {
			model, ok := m.(OpenAPISchemaComponent)
			if ok {
				requestBody[k] = map[string]map[string]any{
					"schema": {
						"$ref": "#/components/schemas/" + model.OpenAPISchemaName(),
					},
				}
				o.AddComponent("schemas", map[string]any{
					model.OpenAPISchemaName(): model.GetOpenAPISchema(),
				})
			} else {
				delete(requestBody, k)
			}
		}
		operationObject["requestBody"] = map[string]any{"content": requestBody}
	}

	responses := op.OpenAPIResponses()
	if responses != nil {
		for code, res := range responses {
			for key, val := range res {
				if key == "content" {
					v, ok := val.(map[string]any)
					if ok {
						for tp, m := range v {
							model, mOK := m.(OpenAPISchemaComponent)
							if mOK {
								responses[code][key] = map[string]map[string]map[string]any{
									tp: {
										"schema": {
											"$ref": "#/components/schemas/" + model.OpenAPISchemaName(),
										},
									},
								}
								o.AddComponent("schemas", map[string]any{
									model.OpenAPISchemaName(): model.GetOpenAPISchema(),
								})
							} else {
								delete(responses[code], key)
							}
						}
					}
				}
			}
		}
		operationObject["responses"] = responses
	}

	if op.OpenAPICallbacks() != nil {
		operationObject["callbacks"] = op.OpenAPICallbacks()
	}
	if op.OpenAPIDeprecated() {
		operationObject["deprecated"] = op.OpenAPIDeprecated()
	}
	if len(op.OpenAPISecurity()) > 0 {
		operationObject["security"] = op.OpenAPISecurity()
	}
	if len(op.OpenAPIServer()) > 0 {
		operationObject["servers"] = op.OpenAPIServer()
	}
	o.AddPath(path, strings.ToLower(method), operationObject)
}

func (o *OpenAPI) Generate(p ...string) error {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	path := "docs/openapi.json"
	if len(p) > 0 {
		path = p[0]
	}
	err = os.WriteFile(path, b, 0666)
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

type OpenAPIInfo struct {
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty"`
	Contact        OpenAPIContact `json:"contact"`
	License        OpenAPILicense `json:"license"`
	Version        string         `json:"version"`
}

type OpenAPIContact struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Url   string `json:"url,omitempty"`
}

type OpenAPILicense struct {
	Name string `json:"name"`
	Url  string `json:"url,omitempty"`
}

type OpenAPIExternalDoc struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url"`
}

// OpenAPIOperationInterface used to complete the operation object of specific endpoint to the open api document
// Full specs is available on https://spec.openapis.org/oas/latest.html#operation-object
type OpenAPIOperationInterface interface {

	// The tags of specific endpoint.
	// for example :
	//	func (Doc) OpenAPITags() []string {
	//		return []string{"Data Store - Product"}
	//	}
	OpenAPITags() []string

	// The summary of specific endpoint.
	// for example :
	//	func (Doc) OpenAPISummary() string {
	//		return "Create Product"
	//	}
	OpenAPISummary() string

	// The description of specific endpoint.
	// for example :
	//	func (Doc) OpenAPISummary() string {
	//		return "Use this method to create contact"
	//	}
	OpenAPIDescription() string

	// The path param of specific endpoint.
	// for example, for endpoint PUT /api/products/{ProductID} you can add the path param like this :
	//	func (Doc) OpenAPIPathParam() []map[string]any {
	//		return []map[string]any{
	//			{
	//				"in":   "path",
	//				"name": "contactID",
	//				"schema": map[string]any{
	//					"type": "string",
	//				},
	//			},
	//		}
	//	}
	// Full specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
	OpenAPIPathParam() []map[string]any

	// The header param of specific endpoint.
	// for example :
	//	func (Doc) OpenAPIHeaderParam() []map[string]any {
	//		return []map[string]any{
	//			{
	//				"in":   "header",
	//				"name": "Content-Language",
	//				"schema": map[string]any{
	//					"type": "string",
	//				},
	//				"examples": map[string]any{
	//					"English (US)": map[string]any{
	//						"value": "en-US",
	//					},
	//					"Bahasa Indonesia": map[string]any{
	//						"value": "id-ID",
	//					},
	//				},
	//			},
	//		}
	//	}
	// Full specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
	OpenAPIHeaderParam() []map[string]any

	// The cookie param of specific endpoint.
	// for example :
	//	func (Doc) OpenAPICookieParam() []map[string]any {
	//		return []map[string]any{
	//			{
	//				"in":   "header",
	//				"name": "token",
	//				"schema": map[string]any{
	//					"type": "string",
	//				},
	//			},
	//		}
	//	}
	// Full specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
	OpenAPICookieParam() []map[string]any

	// The query param of specific endpoint.
	// for example :
	// func (Doc) OpenAPIQueryParam() []map[string]any {
	// 	return []map[string]any{
	// 		{
	// 			"in":   "query",
	// 			"name": "params",
	// 			"schema": map[string]any{
	// 				"type": "object",
	// 				"additionalProperties": map[string]any{
	// 					"type": "string",
	// 				},
	// 			},
	// 			"style":   "form",
	// 			"explode": true,
	// 		},
	// 	}
	// }
	// Full specs is available on https://spec.openapis.org/oas/latest.html#parameter-object
	OpenAPIQueryParam() []map[string]any

	// Body request of specific endpoint.
	// example :
	//	func (Doc) OpenAPIBody() map[string]any {
	//		return map[string]any{
	//			"application/json": &Model{},                       // will auto create schema $ref: '#/components/schemas/Model' if not exists
	//			"application/xml": &Model{},
	//			"application/x-www-form-urlencoded": &Model{},
	//		}
	//	}
	OpenAPIBody() map[string]any

	// Response of specific endpoint.
	// example :
	//	func (Doc) OpenAPIResponses() map[string]map[string]any {
	//		return map[string]map[string]any{
	//			"200": {
	//				"description": "Success",
	//				"content": map[string]any{
	//					"application/json": &Model{},                  // will auto create schema $ref: '#/components/schemas/Model' if not exists
	//					"application/xml": &Model{},
	//				},
	//			},
	//			"401": {
	//				"description": "Unauthorized",
	//				"content": map[string]any{
	//					"application/json": &app.UnauthorizedModel,    // will auto create schema $ref: '#/components/schemas/app.UnauthorizedModel' if not exists
	//					"application/xml": &app.UnauthorizedModel,
	//				},
	//			},
	//		}
	//	}
	// Full specs is available on https://spec.openapis.org/oas/latest.html#response-object
	OpenAPIResponses() map[string]map[string]any

	// Security requirement object of specific endpoint.
	// example :
	//	func (Doc) OpenAPISecurity() []map[string][]string {
	//		return []map[string][]string{
	//			{
	//				"my_app_auth": {
	//					"products:get",
	//					"products:create",
	//				},
	//			},
	//		}
	//	}
	//
	// Full specs is available on https://spec.openapis.org/oas/latest.html#security-requirement-object
	OpenAPISecurity() []map[string][]string

	OpenAPIServer() []map[string]any

	OpenAPIOperationID() string

	OpenAPICallbacks() map[string]any

	OpenAPIDeprecated() bool

	// Allows referencing an external resource for extended documentation of specific endpoint
	// example :
	//	func (Doc) OpenAPIExternalDoc() (string, string) {
	//		return "https://example.com", "Find more info here"
	//	}
	OpenAPIExternalDoc() (string, string)
}

type OpenAPISchemaComponent interface {
	OpenAPISchemaName() string
	GetOpenAPISchema() map[string]any
}

type OpenAPIOperation struct {
	ID              string
	Tags            []string
	Summary         string
	Description     string
	PathParams      []map[string]any
	HeaderParams    []map[string]any
	CookieParams    []map[string]any
	QueryParams     []map[string]any
	Body            map[string]any
	Responses       map[string]map[string]any
	Securities      []map[string][]string
	Servers         []map[string]any
	Callbacks       map[string]any
	IsDeprecated    bool
	ExternalDocUrl  string
	ExternalDocDesc string
}

func (o *OpenAPIOperation) OpenAPITags() []string {
	return o.Tags
}

func (o *OpenAPIOperation) OpenAPISummary() string {
	return o.Summary
}

func (o *OpenAPIOperation) OpenAPIDescription() string {
	return o.Description
}

func (o *OpenAPIOperation) OpenAPIPathParam() []map[string]any {
	return o.PathParams
}

func (o *OpenAPIOperation) OpenAPIHeaderParam() []map[string]any {
	return o.HeaderParams
}

func (o *OpenAPIOperation) OpenAPICookieParam() []map[string]any {
	return o.CookieParams
}

func (o *OpenAPIOperation) OpenAPIQueryParam() []map[string]any {
	return o.QueryParams
}

func (o *OpenAPIOperation) OpenAPIBody() map[string]any {
	return o.Body
}

func (o *OpenAPIOperation) OpenAPIResponses() map[string]map[string]any {
	return o.Responses
}

func (o *OpenAPIOperation) OpenAPISecurity() []map[string][]string {
	return o.Securities
}

func (o *OpenAPIOperation) OpenAPIOperationID() string {
	return o.ID
}

func (o *OpenAPIOperation) OpenAPICallbacks() map[string]any {
	return o.Callbacks
}

func (o *OpenAPIOperation) OpenAPIDeprecated() bool {
	return o.IsDeprecated
}

func (o *OpenAPIOperation) OpenAPIServer() []map[string]any {
	return o.Servers
}

func (o *OpenAPIOperation) OpenAPIExternalDoc() (string, string) {
	return o.ExternalDocUrl, o.ExternalDocDesc
}

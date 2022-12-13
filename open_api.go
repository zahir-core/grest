package grest

type OpenAPIInterface interface {
	SetVersion()
	SetInfo()
	SetJsonSchemaDialect()
	SetServers()
	AddServer(server map[string]any)
	SetTags()
	AddTag(tag map[string]any)
	AddPath(key string, val any)
	SetWebhook()
	AddWebhook(key string, val any)
	AddComponent(key string, val any)
	AddRoute(path, method string, model OpenAPIModelInterface)
	Generate()
}

// The full Latest OpenAPI Specification is available on https://spec.openapis.org/oas/latest.html
type OpenAPI struct {
	OpenAPI           string             `json:"openapi,omitempty"`
	Info              OpenAPIInfo        `json:"info,omitempty"`
	JsonSchemaDialect string             `json:"jsonSchemaDialect,omitempty"`
	Servers           []map[string]any   `json:"servers,omitempty"`
	Paths             map[string]any     `json:"paths,omitempty"`
	Webhooks          map[string]any     `json:"webhooks,omitempty"`
	Components        map[string]any     `json:"components,omitempty"`
	Security          []map[string]any   `json:"security,omitempty"`
	Tags              []map[string]any   `json:"tags,omitempty"`
	ExternalDocs      OpenAPIExternalDoc `json:"externalDocs,omitempty"`
}

func (o *OpenAPI) SetVersion() {
	o.OpenAPI = "3.0.3"
}

func (o *OpenAPI) SetInfo() {
	o.Info.Title = ""
	o.Info.Description = ""
	o.Info.TermsOfService = ""
	o.Info.Contact.Name = ""
	o.Info.Contact.Url = ""
	o.Info.Contact.Email = ""
	o.Info.License.Name = ""
	o.Info.License.Url = ""
	o.Info.Version = ""
	o.ExternalDocs.Url = ""
	o.ExternalDocs.Description = ""
}

func (o *OpenAPI) SetJsonSchemaDialect() {
	o.JsonSchemaDialect = ""
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#server-object
func (o *OpenAPI) SetServers() {
	// example :
	// o.AddServer(map[string]any{
	// 	"url":         "https://localhost:8080",
	// 	"description": "Local Server",
	// })
}

func (o *OpenAPI) AddServer(server map[string]any) {
	o.Servers = append(o.Servers, server)
}

// the full latest specs is available on https://spec.openapis.org/oas/latest.html#tag-object
func (o *OpenAPI) SetTags() {
	// example :
	// o.AddTag(map[string]any{
	// 	"name":        "name",
	// 	"description": "description",
	// })
}

func (o *OpenAPI) AddTag(tag map[string]any) {
	o.Tags = append(o.Tags, tag)
}

func (o *OpenAPI) AddPath(key string, val any) {
	if o.Paths != nil {
		path, isPathExists := o.Paths[key]
		p, pOk := path.(map[string]any)
		v, vOk := val.(map[string]any)
		if isPathExists && pOk && vOk {
			for method, operation := range v {
				_, isMethodExists := p[method]
				if !isMethodExists {
					p[method] = operation
				}
			}
			o.Paths[key] = p
		} else {
			o.Paths[key] = val
		}
	} else {
		o.Paths = map[string]any{key: val}
	}
}

func (o *OpenAPI) SetWebhook() {
	// o.AddWebhook("key", "value")
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

		o.Components[key] = val
	} else {
		o.Components = map[string]any{key: val}
	}
}

func (o *OpenAPI) AddRoute(path, method string, model OpenAPIModelInterface) {
	// todo
}

func (o *OpenAPI) Generate() {
	o.SetVersion()
	o.SetInfo()
	o.SetJsonSchemaDialect()
	o.SetServers()
	o.SetTags()
	o.SetWebhook()
	// todo
}

type OpenAPIInfo struct {
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty"`
	Contact        OpenAPIContact `json:"contact,omitempty"`
	License        OpenAPILicense `json:"license,omitempty"`
	Version        string         `json:"version,omitempty"`
}

type OpenAPIContact struct {
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type OpenAPILicense struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type OpenAPIExternalDoc struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

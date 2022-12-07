package grest

type OpenAPISchema interface {
	// todo
}

// The full Latest OpenAPI Specification is available on https://spec.openapis.org/oas/latest.html
type OpenAPI struct {
	OpenAPI           string             `json:"openapi,omitempty"`
	Info              OpenAPIInfo        `json:"info,omitempty"`
	JsonSchemaDialect string             `json:"jsonSchemaDialect,omitempty"`
	Servers           []OpenAPIServer    `json:"servers,omitempty"`
	Paths             map[string]any     `json:"paths,omitempty"`
	Webhooks          map[string]any     `json:"webhooks,omitempty"`
	Components        map[string]any     `json:"components,omitempty"` // let grest to generate it
	Security          []OpenAPISecurity  `json:"-"`                    // to generate security and components.securitySchemes
	RawSecurity       []any              `json:"security,omitempty"`   // let grest to generate it based on Security
	Tags              []OpenAPITag       `json:"tags,omitempty"`
	ExternalDocs      OpenAPIExternalDoc `json:"externalDocs,omitempty"`
}

func (o *OpenAPI) SetVersion() {
	o.OpenAPI = "3.0.0"
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
}

func (o *OpenAPI) SetJsonSchemaDialect() {
	o.JsonSchemaDialect = ""
}

func (o *OpenAPI) SetServers() {
	// o.AddServer("localhost", "description")
}

func (o *OpenAPI) AddServer(serverUrl, description string) {
	o.Servers = append(o.Servers, OpenAPIServer{Url: serverUrl, Description: description})
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

type OpenAPIServer struct {
	Url         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

type OpenAPITag struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type OpenAPIExternalDoc struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

// example:
// SwaggerSecurity{ID: "basic", Type: "http", Scheme: "basic"}
// SwaggerSecurity{ID: "bearer_token", Type: "http", Scheme: "bearer"}
// SwaggerSecurity{ID: "api_key", Type: "apiKey", Name: "SessionKey", In: "cookie"}
type OpenAPISecurity struct {
	ID               string         `json:"-"`
	Type             string         `json:"type,omitempty"`             // Applies to Any REQUIRED. The type of the security scheme. Valid values are "apiKey", "http", "mutualTLS", "oauth2", "openIdConnect".`
	Description      string         `json:"description,omitempty"`      // Applies to Any A description for security scheme. CommonMark syntax MAY be used for rich text representation.`
	Name             string         `json:"name,omitempty"`             // Applies to apiKey REQUIRED. The name of the header, query or cookie parameter to be used.`
	In               string         `json:"in,omitempty"`               // Applies to apiKey REQUIRED. The location of the API key. Valid values are "query", "header" or "cookie".`
	Scheme           string         `json:"scheme,omitempty"`           // Applies to http REQUIRED. The name of the HTTP Authorization scheme to be used in the Authorization header as defined in RFC7235. The values used SHOULD be registered in the IANA Authentication Scheme registry.`
	BearerFormat     string         `json:"bearerFormat,omitempty"`     // Applies to http ("bearer") A hint to the client to identify how the bearer token is formatted. Bearer tokens are usually generated by an authorization server, so this information is primarily for documentation purposes.`
	Flows            map[string]any `json:"flows,omitempty"`            // Applies to oauth2 REQUIRED. An object containing configuration information for the flow types supported.`
	OpenIdConnectUrl string         `json:"openIdConnectUrl,omitempty"` // Applies to openIdConnect REQUIRED. OpenId Connect URL to discover OAuth2 configuration values. This MUST be in the form of a URL. The OpenID Connect standard requires the use of TLS.`
}

type OpenAPIParam struct {
	Name            string `json:"name,omitempty"`
	In              string `json:"in,omitempty"` // "query", "header", "path" or "cookie".
	Description     string `json:"description,omitempty"`
	Required        bool   `json:"required,omitempty"`
	Deprecated      bool   `json:"deprecated,omitempty"`
	AllowEmptyValue bool   `json:"allowEmptyValue,omitempty"`
}

type OpenAPIRoute struct {
	Path             string             `json:"path,omitempty"`
	Method           string             `json:"method,omitempty"`
	Tags             []string           `json:"tags,omitempty"`
	Summary          string             `json:"summary,omitempty"`
	Description      string             `json:"description,omitempty"`
	OperationId      string             `json:"operationId,omitempty"`
	ExternalDocs     OpenAPIExternalDoc `json:"externalDocs,omitempty"`
	Parameters       []OpenAPIParam     `json:"parameters,omitempty"`
	RequestBody      any                `json:"requestBody,omitempty"`
	RequestBodyType  string             `json:"-"` // used to generate paths, Valid values are "form", "json".`
	RequestBodyModel any                `json:"-"`
	Responses        map[string]any     `json:"responses,omitempty"`
	Security         []any              `json:"security,omitempty"`
}

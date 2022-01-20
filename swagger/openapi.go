// https://swagger.io/specification or https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md
package swagger

type OpenAPI struct {
	OpenAPI           string                 `json:"openapi,omitempty"`
	Info              Info                   `json:"info,omitempty"`
	JsonSchemaDialect string                 `json:"jsonSchemaDialect,omitempty"`
	Servers           []Server               `json:"servers,omitempty"`
	Paths             map[string]interface{} `json:"paths,omitempty"`
	Webhooks          map[string]interface{} `json:"webhooks,omitempty"`
	Components        map[string]interface{} `json:"components,omitempty"` // let grest to generate it
	Security          []Security             `json:"-"`                    // to generate security and components.securitySchemes
	RawSecurity       []interface{}          `json:"security,omitempty"`   // let grest to generate it based on Security
	Tags              []Tag                  `json:"tags,omitempty"`
	ExternalDocs      ExternalDoc            `json:"externalDocs,omitempty"`
}

type Info struct {
	Title          string  `json:"title,omitempty"`
	Description    string  `json:"description,omitempty"`
	TermsOfService string  `json:"termsOfService,omitempty"`
	Contact        Contact `json:"contact,omitempty"`
	License        License `json:"license,omitempty"`
	Version        string  `json:"version,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type Server struct {
	Url         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

type Tag struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type ExternalDoc struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

// example:
// SwaggerSecurity{ID: "basic", Type: "http", Scheme: "basic"}
// SwaggerSecurity{ID: "bearer_token", Type: "http", Scheme: "bearer"}
// SwaggerSecurity{ID: "api_key", Type: "apiKey", Name: "SessionKey", In: "cookie"}
type Security struct {
	ID               string                 `json:"-"`
	Type             string                 `json:"type,omitempty"`             // Applies to Any REQUIRED. The type of the security scheme. Valid values are "apiKey", "http", "mutualTLS", "oauth2", "openIdConnect".`
	Description      string                 `json:"description,omitempty"`      // Applies to Any A description for security scheme. CommonMark syntax MAY be used for rich text representation.`
	Name             string                 `json:"name,omitempty"`             // Applies to apiKey REQUIRED. The name of the header, query or cookie parameter to be used.`
	In               string                 `json:"in,omitempty"`               // Applies to apiKey REQUIRED. The location of the API key. Valid values are "query", "header" or "cookie".`
	Scheme           string                 `json:"scheme,omitempty"`           // Applies to http REQUIRED. The name of the HTTP Authorization scheme to be used in the Authorization header as defined in RFC7235. The values used SHOULD be registered in the IANA Authentication Scheme registry.`
	BearerFormat     string                 `json:"bearerFormat,omitempty"`     // Applies to http ("bearer") A hint to the client to identify how the bearer token is formatted. Bearer tokens are usually generated by an authorization server, so this information is primarily for documentation purposes.`
	Flows            map[string]interface{} `json:"flows,omitempty"`            // Applies to oauth2 REQUIRED. An object containing configuration information for the flow types supported.`
	OpenIdConnectUrl string                 `json:"openIdConnectUrl,omitempty"` // Applies to openIdConnect REQUIRED. OpenId Connect URL to discover OAuth2 configuration values. This MUST be in the form of a URL. The OpenID Connect standard requires the use of TLS.`
}

type Param struct {
	Name            string `json:"name,omitempty"`
	In              string `json:"in,omitempty"` // "query", "header", "path" or "cookie".
	Description     string `json:"description,omitempty"`
	Required        bool   `json:"required,omitempty"`
	Deprecated      bool   `json:"deprecated,omitempty"`
	AllowEmptyValue bool   `json:"allowEmptyValue,omitempty"`
}

type RouteDoc struct {
	Path             string                 `json:"path,omitempty"`
	Method           string                 `json:"method,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	Summary          string                 `json:"summary,omitempty"`
	Description      string                 `json:"description,omitempty"`
	OperationId      string                 `json:"operationId,omitempty"`
	ExternalDocs     ExternalDoc            `json:"externalDocs,omitempty"`
	Parameters       []Param                `json:"parameters,omitempty"`
	RequestBody      interface{}            `json:"requestBody,omitempty"`
	RequestBodyType  string                 `json:"-"` // used to generate paths, Valid values are "form", "json".`
	RequestBodyModel interface{}            `json:"-"`
	Responses        map[string]interface{} `json:"responses,omitempty"`
	Security         []interface{}          `json:"security,omitempty"`
}

func QueryParamDesc() string {
	return `
## Query params

By default, we support a common way for selecting fields, filtering, searching, sorting, and pagination in URL query params on ` + "`GET`" + ` method:

### Field

Get selected fields in GET result, example:
` + "```" + `
GET /api/resources?fields=field_a,field_b,field_c
` + "```" + `
equivalent to sql:
` + "```" + `sql
SELECT field_a, field_b, field_c FROM resources
` + "```" + `

### Filter

Adds fields request condition (multiple conditions) to the request, example:
` + "```" + `
GET /api/resources?field_a=value_a&field_b.$gte=value_b&field_c.$like=value_c&field_d.$ilike=value_d%
` + "```" + `
equivalent to sql:
` + "```" + `sql
SELECT * FROM resources WHERE (field_a = 'value_a') AND (field_b >= value_b) AND (field_c LIKE '%value_c%') AND (LOWER(field_d) LIKE LOWER('value_d%'))
` + "```" + `

#### Available filter conditions

` + "* `$eq`: equal (`=`)" + `
` + "* `$ne`: not equal (`!=`)" + `
` + "* `$gt`: greater than (`>`)" + `
` + "* `$gte`: greater than or equal (`>=`)" + `
` + "* `$lt`: lower than (`<`)" + `
` + "* `$lte`: lower than or equal (`<=`)" + `
` + "* `$like`: contains (`LIKE '%value%'`)" + `
` + "* `$ilike`: contains case insensitive (`LOWER(field) LIKE LOWER('%value%')`)" + `
` + "* `$nlike`: not contains (`NOT LIKE '%value%'`)" + `
` + "* `$nilike`: not contains case insensitive (`LOWER(field) NOT LIKE LOWER('%value%')`)" + `
` + "* `$in`: in range, accepts multiple values (`IN ('value_a', 'value_b')`)" + `
` + "* `$nin`: not in range, accepts multiple values (`NOT IN ('value_a', 'value_b')`)" + `
` + "* `$regexp`: regex (`REGEXP '%value%'`)" + `
` + "* `$nregexp`: not regex (`NOT REGEXP '%value%'`)" + `

### Or

` + "Adds `OR` conditions to the request, example:" + `
` + "```" + `
GET /api/resources?or=field_a:val_a|field_b.$gte:val_b;field_c.$lte:val_c|field_d.$like:val_d
` + "```" + `
equivalent to sql:
` + "```" + `sql
SELECT * FROM resources WHERE (field_a=val_a OR field_b <= val_b) AND (field_c <= val_c OR field_d LIKE '%val_d%')
` + "```" + `

### Search

Adds a search conditions to the request, example:
` + "```" + `
GET /api/resources?search=field_a,field_b:term_1;field_c,field_d:term_2
` + "```" + `
equivalent to sql:
` + "```" + `sql
SELECT * FROM resources WHERE (LOWER(field_a) LIKE LOWER('%term_1%') OR LOWER(field_b) LIKE LOWER('%term_1%')) AND (LOWER(field_c) LIKE LOWER('%term_2%') OR LOWER(field_d) LIKE LOWER('%term_2%'))
` + "```" + `

### Sort

Adds sort by field (by multiple fields) and order to query result, example:
` + "```" + `
GET /api/resources?sorts=field_a,-field_b,field_c:i,-field_d:i
` + "```" + `
equivalent to sql:
` + "```" + `sql
SELECT * FROM resources ORDER BY field_a ASC, field_b DESC, LOWER(field_c) ASC, LOWER(field_d) DESC
` + "```" + `

### Page

Specify the page of results to return, example:
` + "```" + `
GET /api/resources?page=3&per_page=10
` + "```" + `
equivalent to sql:
` + "```" + `sql
SELECT * FROM resources LIMIT 10 OFFSET 20
` + "```" + `

### Per Page

Specify the number of records to return in one request, example:
` + "```" + `
GET /api/resources?per_page=10
` + "```" + `
equivalent to sql:
` + "```" + `sql
SELECT * FROM resources LIMIT 10
` + "```" + `
`
}

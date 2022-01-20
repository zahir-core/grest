package swagger

var (
	swagger OpenAPI
	route   []Route
)

func AddComponent(path, method string, component Component) {
	route = append(route, Route{
		Path:      path,
		Method:    method,
		Component: component,
	})
}

func Generate(openAPI func() OpenAPI) {
	swagger = openAPI()
	if swagger.OpenAPI == "" {
		swagger.OpenAPI = "3.1.0"
	}
	// todo: generate openapi document at docs/openapi.json
}

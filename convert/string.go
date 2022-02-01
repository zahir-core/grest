package convert

import (
	"regexp"
	"strings"
)

func ToCamelCase(str string, d ...string) string {
	delimiter := "_"
	if len(d) > 0 {
		delimiter = d[0]
	}
	link := regexp.MustCompile("(^[A-Za-z])|" + delimiter + "([A-Za-z])")
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, delimiter, "", -1))
	})
}

func ToSnakeCase(str string, d ...string) string {
	delimiter := "_"
	if len(d) > 0 {
		delimiter = d[0]
	}
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}"+delimiter+"${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}"+delimiter+"${2}")
	return strings.ToLower(snake)
}

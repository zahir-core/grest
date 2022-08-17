package grest

import (
	"testing"
)

type strCase struct {
	text, pascalCase, camelCase, snakeCase, kebabCase string
}

func newStringTestData() []strCase {
	sc := []strCase{}
	sc = append(sc, strCase{text: "camel case", pascalCase: "CamelCase", camelCase: "camelCase", snakeCase: "camel_case", kebabCase: "camel-case"})
	sc = append(sc, strCase{text: "snake case", pascalCase: "SnakeCase", camelCase: "snakeCase", snakeCase: "snake_case", kebabCase: "snake-case"})
	sc = append(sc, strCase{text: "kebab case", pascalCase: "KebabCase", camelCase: "kebabCase", snakeCase: "kebab_case", kebabCase: "kebab-case"})
	return sc
}

func TestSnakeCaseToCamelCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := String{}.CamelCase(d.snakeCase)
		if result != d.camelCase {
			t.Errorf("TestSnakeCaseToCamelCase: expected [%v], got [%v]", d.camelCase, result)
		}
	}
}

func TestKebabCaseToCamelCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := String{}.CamelCase(d.kebabCase)
		if result != d.camelCase {
			t.Errorf("TestKebabCaseToCamelCase: expected [%v], got [%v]", d.camelCase, result)
		}
	}
}

func TestCamelCaseToSnakeCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := String{}.SnakeCase(d.camelCase)
		if result != d.snakeCase {
			t.Errorf("TestCamelCaseToSnakeCase: expected [%v], got [%v]", d.snakeCase, result)
		}
	}
}

func TestCamelCaseToKebabCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := String{}.KebabCase(d.camelCase)
		if result != d.kebabCase {
			t.Errorf("TestCamelCaseToKebabCase: expected [%v], got [%v]", d.kebabCase, result)
		}
	}
}

func BenchmarkCamelCaseToCamelCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			String{}.CamelCase(d.camelCase)
		}
	}
}

func BenchmarkSnakeCaseToCamelCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			String{}.CamelCase(d.snakeCase)
		}
	}
}

func BenchmarkKebabCaseToCamelCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			String{}.CamelCase(d.kebabCase)
		}
	}
}

func BenchmarkCamelCaseToSnakeCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			String{}.SnakeCase(d.camelCase)
		}
	}
}

func BenchmarkCamelCaseToKebabCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			String{}.KebabCase(d.camelCase)
		}
	}
}

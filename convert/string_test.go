package convert

import (
	"testing"
)

type strCase struct {
	camelCase, snakeCase, spinalCase string
}

func newStringTestData() []strCase {
	sc := []strCase{}
	sc = append(sc, strCase{camelCase: "CamelCase", snakeCase: "camel_case", spinalCase: "camel-case"})
	sc = append(sc, strCase{camelCase: "SnakeCase", snakeCase: "snake_case", spinalCase: "snake-case"})
	sc = append(sc, strCase{camelCase: "SpinalCase", snakeCase: "spinal_case", spinalCase: "spinal-case"})
	return sc
}

func TestCamelCaseToCamelCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := ToCamelCase(d.camelCase)
		if result != d.camelCase {
			t.Errorf("TestCamelCaseToCamelCase: expected [%v], got [%v]", d.camelCase, result)
		}
	}
}

func TestSnakeCaseToCamelCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := ToCamelCase(d.snakeCase)
		if result != d.camelCase {
			t.Errorf("TestSnakeCaseToCamelCase: expected [%v], got [%v]", d.camelCase, result)
		}
	}
}

func TestSpinalCaseToCamelCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := ToCamelCase(d.spinalCase, "-")
		if result != d.camelCase {
			t.Errorf("TestSpinalCaseToCamelCase: expected [%v], got [%v]", d.camelCase, result)
		}
	}
}

func TestCamelCaseToSnakeCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := ToSnakeCase(d.camelCase)
		if result != d.snakeCase {
			t.Errorf("TestCamelCaseToSnakeCase: expected [%v], got [%v]", d.snakeCase, result)
		}
	}
}

func TestCamelCaseToSpinalCase(t *testing.T) {
	data := newStringTestData()
	for _, d := range data {
		result := ToSnakeCase(d.camelCase, "-")
		if result != d.spinalCase {
			t.Errorf("TestCamelCaseToSpinalCase: expected [%v], got [%v]", d.spinalCase, result)
		}
	}
}

func BenchmarkCamelCaseToCamelCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			ToCamelCase(d.camelCase)
		}
	}
}

func BenchmarkSnakeCaseToCamelCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			ToCamelCase(d.snakeCase)
		}
	}
}

func BenchmarkSpinalCaseToCamelCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			ToCamelCase(d.spinalCase)
		}
	}
}

func BenchmarkCamelCaseToSnakeCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			ToSnakeCase(d.camelCase)
		}
	}
}

func BenchmarkCamelCaseToSpinalCase(b *testing.B) {
	data := newStringTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, d := range data {
			ToSnakeCase(d.camelCase, "-")
		}
	}
}

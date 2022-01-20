package convert

import (
	"testing"
)

func TestConvertToCamelCase(t *testing.T) {
	snake, expected_from_snake := "snake_to_camel", "SnakeToCamel"
	snake_ret := ToCamelCase(snake)
	if snake_ret != expected_from_snake {
		t.Errorf("Expected camel from snake case [%v], got [%v]", expected_from_snake, snake_ret)
	}

	spinal, expected_from_spinal := "spinal-to-camel", "SpinalToCamel"
	spinal_ret := ToCamelCase(spinal, "-")
	if spinal_ret != expected_from_spinal {
		t.Errorf("Expected camel from spinal case [%v], got [%v]", expected_from_spinal, spinal_ret)
	}
}

func TestConvertToSnakeCase(t *testing.T) {
	camel, expected_to_snake := "CamelToSnake", "camel_to_snake"
	snake_ret := ToSnakeCase(camel)
	if snake_ret != expected_to_snake {
		t.Errorf("Expected camel to snake case [%v], got [%v]", expected_to_snake, snake_ret)
	}

	camel, expected_to_spinal := "CamelToSpinal", "camel-to-spinal"
	spinal_ret := ToSnakeCase(camel, "-")
	if spinal_ret != expected_to_spinal {
		t.Errorf("Expected camel to spinal case [%v], got [%v]", expected_to_spinal, spinal_ret)
	}
}

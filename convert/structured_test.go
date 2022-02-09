package convert

import (
	"testing"
)

func TestToStructuredObjectFromByte(t *testing.T) {
	expected := NewDataSample().structuredObject().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().flatObject().String()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestToStructuredObjectFromString(t *testing.T) {
	expected := NewDataSample().structuredObject().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().flatObject().Byte()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}
func TestToStructuredObjectFromAny(t *testing.T) {
	expected := NewDataSample().structuredObject().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().flatObject().Any()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestToStructuredArrayFromByte(t *testing.T) {
	expected := NewDataSample().structuredArray().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().flatArray().String()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestToStructuredArrayFromString(t *testing.T) {
	expected := NewDataSample().structuredArray().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().flatArray().Byte()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}
func TestToStructuredArrayFromAny(t *testing.T) {
	expected := NewDataSample().structuredArray().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().flatArray().Any()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

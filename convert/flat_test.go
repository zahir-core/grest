package convert

import (
	"testing"
)

func TestToFlatObjectFromByte(t *testing.T) {
	expected := NewDataSample().flatObject().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().structuredObject().String()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestToFlatObjectFromString(t *testing.T) {
	expected := NewDataSample().flatObject().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().structuredObject().Byte()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}
func TestToFlatObjectFromAny(t *testing.T) {
	expected := NewDataSample().flatObject().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().structuredObject().Any()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestToFlatArrayFromByte(t *testing.T) {
	expected := NewDataSample().flatArray().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().structuredArray().String()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestToFlatArrayFromString(t *testing.T) {
	expected := NewDataSample().flatArray().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().structuredArray().Byte()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}
func TestToFlatArrayFromAny(t *testing.T) {
	expected := NewDataSample().flatArray().String()
	result := ""
	resultByte, err := NewJSON(NewDataSample().structuredArray().Any()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

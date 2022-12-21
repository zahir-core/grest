package grest

import (
	"encoding/json"
	"testing"
)

type sampleData struct {
	data []byte
	err  error
}

func (d sampleData) Byte() []byte {
	return d.data
}

func (d sampleData) String() string {
	return string(d.data)
}

func (d sampleData) Any() any {
	var v any
	json.Unmarshal(d.data, &v)
	return v
}

func (d sampleData) flatObject() sampleData {
	var v any
	flatRaw := d.flatRaw()
	flat := []byte(flatRaw)
	err := json.Unmarshal(flat, &v)
	if err != nil {
		return sampleData{data: flat, err: err}
	}
	flat, err = json.Marshal(v)
	return sampleData{data: flat, err: err}
}

func (d sampleData) flatArray() sampleData {
	data := "[" + string(d.flatObject().data) + "]"
	return sampleData{data: []byte(data), err: d.err}
}

func (d sampleData) structuredObject() sampleData {
	var v any
	structuredRaw := d.structuredRaw()
	structured := []byte(structuredRaw)
	err := json.Unmarshal(structured, &v)
	if err != nil {
		return sampleData{data: structured, err: err}
	}
	structured, err = json.Marshal(v)
	return sampleData{data: structured, err: err}
}

func (d sampleData) structuredArray() sampleData {
	data := "[" + string(d.structuredObject().data) + "]"
	return sampleData{data: []byte(data), err: d.err}
}

func (sampleData) structuredRaw() string {
	return `{
  "string": "string",
  "bool": true,
  "int": 123,
  "float": 123.45,
  "objek_1": {
    "string": "string",
    "bool": true,
    "number": 123.45,
    "object": {
      "object_1": {
        "string": "string",
        "bool": true
      },
      "object_2": {
        "string": "string",
        "bool": true
      }
    },
    "slice_string": [
      "string_1",
      "string_2",
      "string_3"
    ]
  },
  "objek_2": {
    "string": "string",
    "bool": true,
    "number": 123.45,
    "object": {
      "object_1": {
        "string": "string",
        "bool": true
      },
      "object_2": {
        "string": "string",
        "bool": true,
        "slice_object": [
          {
            "string": "string",
            "object_1": {
              "string": "string",
              "bool": true
            }
          }
        ]
      }
    },
    "objek_2_slice_string": [
      "string_1",
      "string_2",
      "string_3"
    ]
  },
  "slice_string": [
    "string_1",
    "string_2",
    "string_3"
  ],
  "slice_number": [
    1,
    2,
    3
  ],
  "slice_object_1": [
    {
      "string": "string",
      "object_1": {
        "string": "string",
        "bool": true
      },
      "slice_object_1": [
        {
          "string": "string",
          "object_1": {
            "string": "string",
            "bool": true,
            "slice_object_1": [
              {
                "string": "string",
                "object_1": {
                  "string": "string",
                  "bool": true
                }
              }
            ]
          }
        }
      ]
    }
  ]
}`
}

func (sampleData) flatRaw() string {
	return `{
  "bool": true,
  "float": 123.45,
  "int": 123,
  "objek_1.bool": true,
  "objek_1.number": 123.45,
  "objek_1.object.object_1.bool": true,
  "objek_1.object.object_1.string": "string",
  "objek_1.object.object_2.bool": true,
  "objek_1.object.object_2.string": "string",
  "objek_1.slice_string": [
    "string_1",
    "string_2",
    "string_3"
  ],
  "objek_1.string": "string",
  "objek_2.bool": true,
  "objek_2.number": 123.45,
  "objek_2.object.object_1.bool": true,
  "objek_2.object.object_1.string": "string",
  "objek_2.object.object_2.bool": true,
  "objek_2.object.object_2.slice_object": [
    {
      "object_1.bool": true,
      "object_1.string": "string",
      "string": "string"
    }
  ],
  "objek_2.object.object_2.string": "string",
  "objek_2.objek_2_slice_string": [
    "string_1",
    "string_2",
    "string_3"
  ],
  "objek_2.string": "string",
  "slice_number": [
    1,
    2,
    3
  ],
  "slice_object_1": [
    {
      "object_1.bool": true,
      "object_1.string": "string",
      "slice_object_1": [
        {
          "object_1.bool": true,
          "object_1.slice_object_1": [
            {
              "object_1.bool": true,
              "object_1.string": "string",
              "string": "string"
            }
          ],
          "object_1.string": "string",
          "string": "string"
        }
      ],
      "string": "string"
    }
  ],
  "slice_string": [
    "string_1",
    "string_2",
    "string_3"
  ],
  "string": "string"
}
`
}

func TestJsonByteToFlatObject(t *testing.T) {
	expected := sampleData{}.flatObject().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.structuredObject().String()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonStringToFlatObject(t *testing.T) {
	expected := sampleData{}.flatObject().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.structuredObject().Byte()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonAnyToFlatObject(t *testing.T) {
	expected := sampleData{}.flatObject().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.structuredObject().Any()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonByteToFlatArray(t *testing.T) {
	expected := sampleData{}.flatArray().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.structuredArray().String()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonStringToFlatArray(t *testing.T) {
	expected := sampleData{}.flatArray().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.structuredArray().Byte()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonAnyToFlatArray(t *testing.T) {
	expected := sampleData{}.flatArray().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.structuredArray().Any()).ToFlat().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToFlat().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonByteToStructuredObject(t *testing.T) {
	expected := sampleData{}.structuredObject().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.flatObject().String()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonStringToStructuredObject(t *testing.T) {
	expected := sampleData{}.structuredObject().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.flatObject().Byte()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonAnyToStructuredObject(t *testing.T) {
	expected := sampleData{}.structuredObject().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.flatObject().Any()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonByteToStructuredArray(t *testing.T) {
	expected := sampleData{}.structuredArray().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.flatArray().String()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonStringToStructuredArray(t *testing.T) {
	expected := sampleData{}.structuredArray().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.flatArray().Byte()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestJsonAnyToStructuredArray(t *testing.T) {
	expected := sampleData{}.structuredArray().String()
	result := ""
	resultByte, err := NewJSON(sampleData{}.flatArray().Any()).ToStructured().Marshal()
	if err == nil {
		result = string(resultByte)
	} else {
		t.Errorf("NewJSON().ToStructured().Marshal() error [%v]", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func BenchmarkJsonByteToFlatObject(b *testing.B) {
	jsonByte := sampleData{}.structuredObject().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToFlat()
	}
}

func BenchmarkJsonByteToFlatObjectMarshal(b *testing.B) {
	jsonByte := sampleData{}.structuredObject().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToFlat().Marshal()
	}
}

func BenchmarkJsonByteToFlatObjectUnmarshal(b *testing.B) {
	jsonByte := sampleData{}.structuredObject().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := map[string]any{}
		NewJSON(jsonByte).ToFlat().Unmarshal(&v)
	}
}

func BenchmarkJsonStringToFlatObject(b *testing.B) {
	jsonString := sampleData{}.structuredObject().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToFlat()
	}
}

func BenchmarkJsonStringToFlatObjectMarshal(b *testing.B) {
	jsonString := sampleData{}.structuredObject().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToFlat().Marshal()
	}
}

func BenchmarkJsonStringToFlatObjectUnmarshal(b *testing.B) {
	jsonString := sampleData{}.structuredObject().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := map[string]any{}
		NewJSON(jsonString).ToFlat().Unmarshal(&v)
	}
}

func BenchmarkJsonAnyToFlatObject(b *testing.B) {
	jsonAny := sampleData{}.structuredObject().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToFlat()
	}
}

func BenchmarkJsonAnyToFlatObjectMarshal(b *testing.B) {
	jsonAny := sampleData{}.structuredObject().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToFlat().Marshal()
	}
}

func BenchmarkJsonAnyToFlatObjectUnmarshal(b *testing.B) {
	jsonAny := sampleData{}.structuredObject().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := map[string]any{}
		NewJSON(jsonAny).ToFlat().Unmarshal(&v)
	}
}

func BenchmarkJsonByteToStructuredObject(b *testing.B) {
	jsonByte := sampleData{}.flatObject().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToStructured()
	}
}

func BenchmarkJsonByteToStructuredObjectMarshal(b *testing.B) {
	jsonByte := sampleData{}.flatObject().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToStructured().Marshal()
	}
}

func BenchmarkJsonByteToStructuredObjectUnmarshal(b *testing.B) {
	jsonByte := sampleData{}.flatObject().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := map[string]any{}
		NewJSON(jsonByte).ToStructured().Unmarshal(&v)
	}
}

func BenchmarkJsonStringToStructuredObject(b *testing.B) {
	jsonString := sampleData{}.flatObject().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToStructured()
	}
}

func BenchmarkJsonStringToStructuredObjectMarshal(b *testing.B) {
	jsonString := sampleData{}.flatObject().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToStructured().Marshal()
	}
}

func BenchmarkJsonStringToStructuredObjectUnmarshal(b *testing.B) {
	jsonString := sampleData{}.flatObject().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := map[string]any{}
		NewJSON(jsonString).ToStructured().Unmarshal(&v)
	}
}

func BenchmarkJsonAnyToStructuredObject(b *testing.B) {
	jsonAny := sampleData{}.flatObject().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToStructured()
	}
}

func BenchmarkJsonAnyToStructuredObjectMarshal(b *testing.B) {
	jsonAny := sampleData{}.flatObject().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToStructured().Marshal()
	}
}

func BenchmarkJsonAnyToStructuredObjectUnmarshal(b *testing.B) {
	jsonAny := sampleData{}.flatObject().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := map[string]any{}
		NewJSON(jsonAny).ToStructured().Unmarshal(&v)
	}
}

func BenchmarkJsonByteToFlatArray(b *testing.B) {
	jsonByte := sampleData{}.structuredArray().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToFlat()
	}
}

func BenchmarkJsonByteToFlatArrayMarshal(b *testing.B) {
	jsonByte := sampleData{}.structuredArray().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToFlat().Marshal()
	}
}

func BenchmarkJsonByteToFlatArrayUnmarshal(b *testing.B) {
	jsonByte := sampleData{}.structuredArray().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := []map[string]any{}
		NewJSON(jsonByte).ToFlat().Unmarshal(&v)
	}
}

func BenchmarkJsonStringToFlatArray(b *testing.B) {
	jsonString := sampleData{}.structuredArray().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToFlat()
	}
}

func BenchmarkJsonStringToFlatArrayMarshal(b *testing.B) {
	jsonString := sampleData{}.structuredArray().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToFlat().Marshal()
	}
}

func BenchmarkJsonStringToFlatArrayUnmarshal(b *testing.B) {
	jsonString := sampleData{}.structuredArray().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := []map[string]any{}
		NewJSON(jsonString).ToFlat().Unmarshal(&v)
	}
}

func BenchmarkJsonAnyToFlatArray(b *testing.B) {
	jsonAny := sampleData{}.structuredArray().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToFlat()
	}
}

func BenchmarkJsonAnyToFlatArrayMarshal(b *testing.B) {
	jsonAny := sampleData{}.structuredArray().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToFlat().Marshal()
	}
}

func BenchmarkJsonAnyToFlatArrayUnmarshal(b *testing.B) {
	jsonAny := sampleData{}.structuredArray().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := []map[string]any{}
		NewJSON(jsonAny).ToFlat().Unmarshal(&v)
	}
}

func BenchmarkJsonByteToStructuredArray(b *testing.B) {
	jsonByte := sampleData{}.flatArray().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToStructured()
	}
}

func BenchmarkJsonByteToStructuredArrayMarshal(b *testing.B) {
	jsonByte := sampleData{}.flatArray().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonByte).ToStructured().Marshal()
	}
}

func BenchmarkJsonByteToStructuredArrayUnmarshal(b *testing.B) {
	jsonByte := sampleData{}.flatArray().Byte()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := []map[string]any{}
		NewJSON(jsonByte).ToStructured().Unmarshal(&v)
	}
}

func BenchmarkJsonStringToStructuredArray(b *testing.B) {
	jsonString := sampleData{}.flatArray().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToStructured()
	}
}

func BenchmarkJsonStringToStructuredArrayMarshal(b *testing.B) {
	jsonString := sampleData{}.flatArray().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonString).ToStructured().Marshal()
	}
}

func BenchmarkJsonStringToStructuredArrayUnmarshal(b *testing.B) {
	jsonString := sampleData{}.flatArray().String()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := []map[string]any{}
		NewJSON(jsonString).ToStructured().Unmarshal(&v)
	}
}

func BenchmarkJsonAnyToStructuredArray(b *testing.B) {
	jsonAny := sampleData{}.flatArray().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToStructured()
	}
}

func BenchmarkJsonAnyToStructuredArrayMarshal(b *testing.B) {
	jsonAny := sampleData{}.flatArray().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewJSON(jsonAny).ToStructured().Marshal()
	}
}

func BenchmarkJsonAnyToStructuredArrayUnmarshal(b *testing.B) {
	jsonAny := sampleData{}.flatArray().Any()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := []map[string]any{}
		NewJSON(jsonAny).ToStructured().Unmarshal(&v)
	}
}

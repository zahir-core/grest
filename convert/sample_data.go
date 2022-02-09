package convert

import "encoding/json"

type dataSample struct {
	data []byte
	err  error
}

func NewDataSample() dataSample {
	return dataSample{}
}

func (d dataSample) Byte() []byte {
	return d.data
}

func (d dataSample) String() string {
	return string(d.data)
}

func (d dataSample) Any() interface{} {
	var v interface{}
	json.Unmarshal(d.data, &v)
	return v
}

func (d dataSample) flatObject() dataSample {
	var v interface{}
	flatRaw := d.flatRaw()
	flat := []byte(flatRaw)
	err := json.Unmarshal(flat, &v)
	if err != nil {
		return dataSample{data: flat, err: err}
	}
	flat, err = json.Marshal(v)
	return dataSample{data: flat, err: err}
}

func (d dataSample) flatArray() dataSample {
	data := "[" + string(d.flatObject().data) + "]"
	return dataSample{data: []byte(data), err: d.err}
}

func (d dataSample) structuredObject() dataSample {
	var v interface{}
	structuredRaw := d.structuredRaw()
	structured := []byte(structuredRaw)
	err := json.Unmarshal(structured, &v)
	if err != nil {
		return dataSample{data: structured, err: err}
	}
	structured, err = json.Marshal(v)
	return dataSample{data: structured, err: err}
}

func (d dataSample) structuredArray() dataSample {
	data := "[" + string(d.structuredObject().data) + "]"
	return dataSample{data: []byte(data), err: d.err}
}

func (dataSample) structuredRaw() string {
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

func (dataSample) flatRaw() string {
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

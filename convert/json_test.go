package convert

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

type friend struct {
	FirstName string `json:"name.first,omitempty"`
	LastName  string `json:"name.last,omitempty"`
}

type person struct {
	ID        string    `json:"id,omitempty"`
	FirstName string    `json:"name.first,omitempty"`
	LastName  string    `json:"name.last,omitempty"`
	Age       int       `json:"age,omitempty"`
	Friends   []friend  `json:"friends"`
	CreatedAt time.Time `json:"created.time,omitempty"`
	UpdatedAt time.Time `json:"updated.time,omitempty"`
}

func newJsonTestData() (expected person, flatJsonObject, flatJsonArray, structuredJsonObject, structuredJsonArray []byte) {
	e := person{}
	e.ID = uuid.NewString()
	e.FirstName = "Phay"
	e.LastName = "Joe"
	e.Age = rand.Intn(100)
	e.Friends = append(e.Friends, friend{FirstName: "John", LastName: "Thor"})
	e.Friends = append(e.Friends, friend{FirstName: "Ryan", LastName: "Tho"})
	e.CreatedAt, _ = time.Parse(time.RFC3339, "2021-08-30T10:12:09Z")
	e.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	flatJsonObjectString := `{
		"id": "` + e.ID + `",
		"name.first": "` + e.FirstName + `",
		"name.last": "` + e.LastName + `",
		"age": "` + strconv.Itoa(e.Age) + `",
		"friends": [
			{
				"name.first": "` + e.Friends[0].FirstName + `",
				"name.last": "` + e.Friends[0].LastName + `"
			},
			{
				"name.first": "` + e.Friends[1].FirstName + `",
				"name.last": "` + e.Friends[1].LastName + `"
			}
		],
		"created.time": "` + e.CreatedAt.Format(time.RFC3339) + `",
		"updated.time": "` + e.UpdatedAt.Format(time.RFC3339) + `"
	}`

	structuredJsonObjectString := `{
		"id": "` + e.ID + `",
		"name": {
			"first": "` + e.FirstName + `",
			"last": "` + e.LastName + `"
		},
		"age": "` + strconv.Itoa(e.Age) + `",
		"friends": [
			{
				"name": {
					"first": "` + e.Friends[0].FirstName + `",
					"last": "` + e.Friends[0].LastName + `"
				}
			},
			{
				"name": {
					"first": "` + e.Friends[1].FirstName + `",
					"last": "` + e.Friends[1].LastName + `"
				}
			}
		],
		"created": {
			"time": "` + e.CreatedAt.Format(time.RFC3339) + `"
		},
		"updated": {
			"time": "` + e.UpdatedAt.Format(time.RFC3339) + `"
		}
	}`

	expected = e
	flatJsonObject = []byte(flatJsonObjectString)
	flatJsonArray = []byte("[" + flatJsonObjectString + "]")
	structuredJsonObject = []byte(structuredJsonObjectString)
	structuredJsonArray = []byte("[" + structuredJsonObjectString + "]")
	return
}

func checkUnmarshalTestData(t *testing.T, title string, expected, result person) {
	if result.ID != expected.ID {
		t.Errorf("%v : Expected ID [%v], got [%v]", title, expected.ID, result.ID)
	}
	if result.FirstName != expected.FirstName {
		t.Errorf("%v : Expected FirstName [%v], got [%v]", title, expected.FirstName, result.FirstName)
	}
	if result.LastName != expected.LastName {
		t.Errorf("%v : Expected LastName [%v], got [%v]", title, expected.LastName, result.LastName)
	}
	if result.Age != expected.Age {
		t.Errorf("%v : Expected Age [%v], got [%v]", title, expected.Age, result.Age)
	}
	if len(result.Friends) != len(expected.Friends) {
		t.Errorf("%v : Expected Friends count [%v], got [%v]", title, len(expected.Friends), len(result.Friends))
	} else {
		if result.Friends[0].FirstName != expected.Friends[0].FirstName {
			t.Errorf("%v : Expected Friends[0].FirstName [%v], got [%v]", title, expected.Friends[0].FirstName, result.Friends[0].FirstName)
		}
		if result.Friends[0].LastName != expected.Friends[0].LastName {
			t.Errorf("%v : Expected Friends[0].LastName [%v], got [%v]", title, expected.Friends[0].LastName, result.Friends[0].LastName)
		}
		if result.Friends[1].FirstName != expected.Friends[1].FirstName {
			t.Errorf("%v : Expected Friends[1].FirstName [%v], got [%v]", title, expected.Friends[1].FirstName, result.Friends[1].FirstName)
		}
		if result.Friends[1].LastName != expected.Friends[1].LastName {
			t.Errorf("%v : Expected Friends[1].LastName [%v], got [%v]", title, expected.Friends[1].LastName, result.Friends[1].LastName)
		}
	}
	if result.CreatedAt != expected.CreatedAt {
		t.Errorf("%v : Expected CreatedAt [%v], got [%v]", title, expected.CreatedAt, result.CreatedAt)
	}
	if result.UpdatedAt != expected.UpdatedAt {
		t.Errorf("%v : Expected UpdatedAt [%v], got [%v]", title, expected.UpdatedAt, result.UpdatedAt)
	}
}

func TestUnmarshalFlatJSONFromStructuredJSONObject(t *testing.T) {
	expected, _, _, structuredJsonObject, _ := newJsonTestData()
	result := person{}
	err := ToFlatJSON(structuredJsonObject, result).Unmarshal(&result)
	if err != nil {
		t.Errorf("Test unmarshal flat JSON from structured JSON Object : Error occurred [%v]", err)
	}
	checkUnmarshalTestData(t, "Test unmarshal flat JSON from structured JSON Object", expected, result)
}

func TestUnmarshalFlatJSONFromStructuredJSONArray(t *testing.T) {
	expected, _, _, _, structuredJsonArray := newJsonTestData()
	result := []person{}
	err := ToFlatJSON(structuredJsonArray, result).Unmarshal(&result)
	if err != nil {
		t.Errorf("Test unmarshal flat JSON from structured JSON Array : Error occurred [%v]", err)
	}
	if len(result) == 0 {
		t.Errorf("Test unmarshal flat JSON from structured JSON Array : Expected count [%v], got [%v]", 1, 0)
	} else {
		checkUnmarshalTestData(t, "Test unmarshal flat JSON from structured JSON Array", expected, result[0])
	}
}

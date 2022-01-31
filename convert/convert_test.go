package convert

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
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

func TestToFlatJSON(t *testing.T) {
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
	}

	e := person{}
	e.ID = "some-uuid"
	e.FirstName = "Janet"
	e.LastName = "Prichard"
	e.Age = 25
	e.Friends = append(e.Friends, friend{FirstName: "John", LastName: "Thor"})
	e.Friends = append(e.Friends, friend{FirstName: "Ryan", LastName: "Tho"})
	e.CreatedAt, _ = time.Parse(time.RFC3339, "2021-08-30T10:12:09Z")

	jsonObject := `{"id":"` + e.ID + `","age":"` + strconv.Itoa(e.Age) + `","created":{"time":"` + e.CreatedAt.Format(time.RFC3339) + `"},"name":{"first":"` + e.FirstName + `","last":"` + e.LastName + `"},"friends":[{"name":{"first":"` + e.Friends[0].FirstName + `","last":"` + e.Friends[0].LastName + `"}},{"name":{"first":"` + e.Friends[1].FirstName + `","last":"` + e.Friends[1].LastName + `"}}]}`

	p := person{}
	flatJSONObject, err := ToFlatJSON(p, []byte(jsonObject))
	if err != nil {
		t.Errorf("ToFlatJSON: Error occurred [%v]", err)
	}
	if err := json.Unmarshal(flatJSONObject, &p); err != nil {
		t.Errorf("json.Unmarshal: Error occurred [%v]", err)
	}

	if p.ID != e.ID {
		t.Errorf("Expected jsonObject ID [%v], got [%v]", e.ID, p.ID)
	}
	if p.FirstName != e.FirstName {
		t.Errorf("Expected jsonObject FirstName [%v], got [%v]", e.FirstName, p.FirstName)
	}
	if p.LastName != e.LastName {
		t.Errorf("Expected jsonObject LastName [%v], got [%v]", e.LastName, p.LastName)
	}
	if p.Age != e.Age {
		t.Errorf("Expected jsonObject Age [%v], got [%v]", e.Age, p.Age)
	}
	if len(p.Friends) != len(e.Friends) {
		t.Errorf("Expected jsonObject Friends count [%v], got [%v]", len(e.Friends), len(p.Friends))
	}
	if p.Friends[0].FirstName != e.Friends[0].FirstName {
		t.Errorf("Expected jsonObject Friends[0].FirstName [%v], got [%v]", e.Friends[0].FirstName, p.Friends[0].FirstName)
	}
	if p.Friends[0].LastName != e.Friends[0].LastName {
		t.Errorf("Expected jsonObject Friends[0].LastName [%v], got [%v]", e.Friends[0].LastName, p.Friends[0].LastName)
	}
	if p.Friends[1].FirstName != e.Friends[1].FirstName {
		t.Errorf("Expected jsonObject Friends[1].FirstName [%v], got [%v]", e.Friends[1].FirstName, p.Friends[1].FirstName)
	}
	if p.Friends[1].LastName != e.Friends[1].LastName {
		t.Errorf("Expected jsonObject Friends[1].LastName [%v], got [%v]", e.Friends[1].LastName, p.Friends[1].LastName)
	}
	if p.CreatedAt != e.CreatedAt {
		t.Errorf("Expected jsonObject CreatedAt [%v], got [%v]", e.CreatedAt, p.CreatedAt)
	}

	jsonArray := `[` + jsonObject + `]`

	a := []person{}
	flatJSONArray, err := ToFlatJSON(a, []byte(jsonArray))
	if err != nil {
		t.Errorf("ToFlatJSON: Error occurred [%v]", err)
	}
	if err := json.Unmarshal(flatJSONArray, &a); err != nil {
		t.Errorf("json.Unmarshal: Error occurred [%v]", err)
	}

	if a[0].ID != e.ID {
		t.Errorf("Expected jsonArray ID [%v], got [%v]", e.ID, a[0].ID)
	}
	if a[0].FirstName != e.FirstName {
		t.Errorf("Expected jsonArray FirstName [%v], got [%v]", e.FirstName, a[0].FirstName)
	}
	if a[0].LastName != e.LastName {
		t.Errorf("Expected jsonArray LastName [%v], got [%v]", e.LastName, a[0].LastName)
	}
	if a[0].Age != e.Age {
		t.Errorf("Expected jsonArray Age [%v], got [%v]", e.Age, a[0].Age)
	}
	if len(a[0].Friends) != len(e.Friends) {
		t.Errorf("Expected jsonArray Friends count [%v], got [%v]", len(e.Friends), len(a[0].Friends))
	}
	if a[0].Friends[0].FirstName != e.Friends[0].FirstName {
		t.Errorf("Expected jsonArray Friends[0].FirstName [%v], got [%v]", e.Friends[0].FirstName, a[0].Friends[0].FirstName)
	}
	if a[0].Friends[0].LastName != e.Friends[0].LastName {
		t.Errorf("Expected jsonArray Friends[0].LastName [%v], got [%v]", e.Friends[0].LastName, a[0].Friends[0].LastName)
	}
	if a[0].Friends[1].FirstName != e.Friends[1].FirstName {
		t.Errorf("Expected jsonArray Friends[1].FirstName [%v], got [%v]", e.Friends[1].FirstName, a[0].Friends[1].FirstName)
	}
	if a[0].Friends[1].LastName != e.Friends[1].LastName {
		t.Errorf("Expected jsonArray Friends[1].LastName [%v], got [%v]", e.Friends[1].LastName, a[0].Friends[1].LastName)
	}
	if a[0].CreatedAt != e.CreatedAt {
		t.Errorf("Expected jsonArray CreatedAt [%v], got [%v]", e.CreatedAt, a[0].CreatedAt)
	}
}

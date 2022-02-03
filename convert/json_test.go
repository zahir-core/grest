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
		"age": ` + strconv.Itoa(e.Age) + `,
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
		"age": ` + strconv.Itoa(e.Age) + `,
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

func checkUnmarshalStructTestData(t *testing.T, title string, expected, result person) {
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

func checkUnmarshalMapStructuredTestData(t *testing.T, title string, expected person, result map[string]interface{}) {
	id, isIdExists := result["id"]
	if !isIdExists {
		t.Errorf("%v : id is not exists", title)
	}
	idString, isIdString := id.(string)
	if !isIdString {
		t.Errorf("%v : id is not string", title)
	}
	if idString != expected.ID {
		t.Errorf("%v : expected ID [%v], got [%v]", title, expected.ID, idString)
	}
	name, isNameExists := result["name"]
	if !isNameExists {
		t.Errorf("%v : name is not exists", title)
	}
	nameMap, isNameMap := name.(map[string]interface{})
	if !isNameMap {
		t.Errorf("%v : name is not map", title)
	}
	firstName, isFirstNameExists := nameMap["first"]
	if !isFirstNameExists {
		t.Errorf("%v : name[first] is not exists", title)
	}
	firstNameString, isFirstNameString := firstName.(string)
	if !isFirstNameString {
		t.Errorf("%v : name[first] is not string", title)
	}
	if firstNameString != expected.FirstName {
		t.Errorf("%v : expected name[first] [%v], got [%v]", title, expected.FirstName, firstNameString)
	}
	lastName, isLastNameExists := nameMap["last"]
	if !isLastNameExists {
		t.Errorf("%v : name[last] is not exists", title)
	}
	lastNameString, isLastNameString := lastName.(string)
	if !isLastNameString {
		t.Errorf("%v : name[last] is not string", title)
	}
	if lastNameString != expected.LastName {
		t.Errorf("%v : expected name[last] [%v], got [%v]", title, expected.LastName, lastNameString)
	}
	age, isAgeExists := result["age"]
	if !isAgeExists {
		t.Errorf("%v : age is not exists", title)
	}
	ageInt, isAgeInt := age.(float64)
	if !isAgeInt {
		t.Errorf("%v : age is not float64", title)
	}
	if int(ageInt) != expected.Age {
		t.Errorf("%v : expected age [%v], got [%v]", title, expected.Age, ageInt)
	}
	friends, isFriendsExists := result["friends"]
	if !isFriendsExists {
		t.Errorf("%v : friends is not exists", title)
	}
	friendsSlice, isFriendsSlice := friends.([]interface{})
	if !isFriendsSlice {
		t.Errorf("%v : friends is not slice", title)
	}
	if len(friendsSlice) != len(expected.Friends) {
		t.Errorf("%v : expected friends count [%v], got [%v]", title, len(expected.Friends), len(friendsSlice))
	} else {
		for i, friendsSliceValue := range friendsSlice {
			v, isFriefriendsSliceValueMap := friendsSliceValue.(map[string]interface{})
			if !isFriefriendsSliceValueMap {
				t.Errorf("%v : friends slice value is not map", title)
			}
			friendName, isFriendNameExists := v["name"]
			if !isFriendNameExists {
				t.Errorf("%v : friends[%v].name is not exists", title, i)
			}
			friendNameMap, isFriendNameMap := friendName.(map[string]interface{})
			if !isFriendNameMap {
				t.Errorf("%v : friends[%v].name is not map", title, i)
			}
			friendFirstName, isFriendFirstNameExists := friendNameMap["first"]
			if !isFriendFirstNameExists {
				t.Errorf("%v : friends[%v].name[first] is not exists", title, i)
			}
			if friendFirstName != expected.Friends[i].FirstName {
				t.Errorf("%v : expected friends[%v].name[first] [%v], got [%v]", title, i, expected.Friends[i].FirstName, friendFirstName)
			}
			friendLastName, isFriendLastNameExists := friendNameMap["last"]
			if !isFriendLastNameExists {
				t.Errorf("%v : friends[%v].name[last] is not exists", title, i)
			}
			if friendLastName != expected.Friends[i].LastName {
				t.Errorf("%v : expected friends[%v].name[last] [%v], got [%v]", title, i, expected.Friends[i].LastName, friendLastName)
			}

		}
	}
	created, isCreatedExists := result["created"]
	if !isCreatedExists {
		t.Errorf("%v : created is not exists", title)
	}
	createdMap, isCreatedMap := created.(map[string]interface{})
	if !isCreatedMap {
		t.Errorf("%v : created is not map", title)
	}
	createdTime, isCreatedTimeExists := createdMap["time"]
	if !isCreatedTimeExists {
		t.Errorf("%v : created[time] is not exists", title)
	}
	createdTimeTime, createdTimeErr := time.Parse(time.RFC3339, createdTime.(string))
	if createdTimeErr != nil {
		t.Errorf("%v : created[time] is not time.Time, err :%v", title, createdTimeErr.Error())
	}
	if createdTimeTime != expected.CreatedAt {
		t.Errorf("%v : expected created[time] [%v], got [%v]", title, expected.CreatedAt, createdTimeTime)
	}
	updated, isUpdatedExists := result["updated"]
	if !isUpdatedExists {
		t.Errorf("%v : updated is not exists", title)
	}
	updatedMap, isUpdatedMap := updated.(map[string]interface{})
	if !isUpdatedMap {
		t.Errorf("%v : updated is not map", title)
	}
	updatedTime, isUpdatedTimeExists := updatedMap["time"]
	if !isUpdatedTimeExists {
		t.Errorf("%v : updated[time] is not exists", title)
	}
	updatedTimeTime, updatedTimeErr := time.Parse(time.RFC3339, updatedTime.(string))
	if updatedTimeErr != nil {
		t.Errorf("%v : updated[time] is not time.Time, err :%v", title, updatedTimeErr.Error())
	}
	if updatedTimeTime != expected.UpdatedAt {
		t.Errorf("%v : expected updated[time] [%v], got [%v]", title, expected.UpdatedAt, updatedTimeTime)
	}
}

func TestUnmarshalFlatJSONFromStructuredJSONObjectBasedOnStrucType(t *testing.T) {
	expected, _, _, structuredJsonObject, _ := newJsonTestData()
	result := person{}
	err := ToFlatJSON(structuredJsonObject, result).Unmarshal(&result)
	if err != nil {
		t.Errorf("Test unmarshal flat JSON from structured JSON Object based on struct type : Error occurred [%v]", err)
	}
	checkUnmarshalStructTestData(t, "Test unmarshal flat JSON from structured JSON Object based on struct type", expected, result)
}

func TestUnmarshalFlatJSONFromStructuredJSONArrayBasedOnStrucType(t *testing.T) {
	expected, _, _, _, structuredJsonArray := newJsonTestData()
	result := []person{}
	err := ToFlatJSON(structuredJsonArray, result).Unmarshal(&result)
	if err != nil {
		t.Errorf("Test unmarshal flat JSON from structured JSON Array based on struct type : Error occurred [%v]", err)
	}
	if len(result) == 0 {
		t.Errorf("Test unmarshal flat JSON from structured JSON Array based on struct type : Expected count [%v], got [%v]", 1, 0)
	} else {
		checkUnmarshalStructTestData(t, "Test unmarshal flat JSON from structured JSON Array based on struct type", expected, result[0])
	}
}

func TestUnmarshalStructuredJSONFromFlatJSONObjectByte(t *testing.T) {
	expected, flatJsonObject, _, _, _ := newJsonTestData()
	result := map[string]interface{}{}
	err := ToStructuredJSON(flatJsonObject, nil).Unmarshal(&result)
	if err != nil {
		t.Errorf("Test unmarshal structured JSON from flat JSON object byte : Error occurred [%v]", err)
	}
	checkUnmarshalMapStructuredTestData(t, "Test unmarshal structured JSON from flat JSON object byte", expected, result)
}

func BenchmarkUnmarshalFlatJSONFromStructuredJSONObject(b *testing.B) {
	_, _, _, structuredJsonObject, _ := newJsonTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := person{}
		ToFlatJSON(structuredJsonObject, result).Unmarshal(&result)
	}
}

func BenchmarkMarshalFlatJSONFromStructuredJSONObject(b *testing.B) {
	_, _, _, structuredJsonObject, _ := newJsonTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := person{}
		ToFlatJSON(structuredJsonObject, result).Marshal()
	}
}

func BenchmarkUnmarshalFlatJSONFromStructuredJSONArray(b *testing.B) {
	_, _, _, _, structuredJsonArray := newJsonTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := []person{}
		ToFlatJSON(structuredJsonArray, result).Unmarshal(&result)
	}
}

func BenchmarkMarshalFlatJSONFromStructuredJSONArray(b *testing.B) {
	_, _, _, _, structuredJsonArray := newJsonTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := []person{}
		ToFlatJSON(structuredJsonArray, result).Marshal()
	}
}

func BenchmarkUnmarshalStructuredJSONFromFlatJSONObjectByte(b *testing.B) {
	_, flatJsonObject, _, _, _ := newJsonTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := map[string]interface{}{}
		ToStructuredJSON(flatJsonObject, nil).Unmarshal(&result)
	}
}

func BenchmarkMarshalStructuredJSONFromFlatJSONObjectByte(b *testing.B) {
	_, flatJsonObject, _, _, _ := newJsonTestData()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ToStructuredJSON(flatJsonObject, nil).Marshal()
	}
}

package grest

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestHttpClientSimple(t *testing.T) {
	expected := []byte("OK")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(expected)
	}))
	defer server.Close()

	c := NewHttpClient("GET", server.URL)
	res, err := c.Send()
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusCode [%v], got [%v]", http.StatusOK, res.StatusCode)
	}
	if !reflect.DeepEqual(expected, c.BodyResponse) {
		t.Errorf("Expected BodyResponse [%v], got [%v]", expected, c.BodyResponse)
	}
}

func TestHttpClientWithHeaderRequest(t *testing.T) {
	headerKey, headerValue := "header-key", "header-value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		hd := req.Header.Get(headerKey)
		if hd != headerValue {
			t.Errorf("Expected headerValue [%v], got [%v]", headerValue, hd)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	c := NewHttpClient("GET", server.URL)
	c.AddHeader(headerKey, headerValue)
	_, err := c.Send()
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func TestHttpClientWithMultipartBodyRequest(t *testing.T) {
	type bodyRequest struct {
		Name  string   `form:"name"`
		Image *os.File `form:"image"`
	}

	expectedName := "some_value"
	filename := "zahir-logo.jpg"
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Error occurred when open file [%v]", err)
	} else {

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			name := req.FormValue("name")
			if name != expectedName {
				t.Errorf("Expected name [%v], got [%v]", expectedName, name)
			}
			_, fh, err := req.FormFile("image")
			if err != nil {
				t.Errorf("Error occurred FormFile [%v]", err)
			} else {
				if fh.Filename != filename {
					t.Errorf("Expected filename [%v], got [%v]", filename, fh.Filename)
				}
			}
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		c := NewHttpClient("POST", server.URL)
		c.AddMultipartBody(bodyRequest{
			Name:  expectedName,
			Image: file,
		})
		_, err := c.Send()
		if err != nil {
			t.Errorf("Error occurred [%v]", err)
		}
	}
}

func TestHttpClientWithUrlEncodedBodyRequest(t *testing.T) {
	type bodyRequest struct {
		Name string `form:"name"`
	}
	expectedName := "some_value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		expectedContentType := "application/x-www-form-urlencoded"
		contentType := req.Header.Get("Content-Type")
		if expectedContentType != contentType {
			t.Errorf("Expected Content-Type [%v], got [%v]", expectedContentType, contentType)
		}
		name := req.FormValue("name")
		if name != expectedName {
			t.Errorf("Expected name [%v], got [%v]", expectedName, name)
		}
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	c := NewHttpClient("POST", server.URL)
	c.AddUrlEncodedBody(bodyRequest{
		Name: expectedName,
	})
	_, err := c.Send()
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func TestHttpClientWithJsonBodyRequest(t *testing.T) {
	type bodyRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	expectedName := "some_value"
	expectedAge := 23

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		expectedContentType := "application/json"
		contentType := req.Header.Get("Content-Type")
		if expectedContentType != contentType {
			t.Errorf("Expected Content-Type [%v], got [%v]", expectedContentType, contentType)
		}

		body, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			t.Errorf("Error occurred when read body [%v]", err)
		} else {
			b := bodyRequest{}
			err = json.Unmarshal(body, &b)
			if err != nil {
				t.Errorf("Error occurred when unmarshal json [%v]", err)
			}
			if b.Name != expectedName {
				t.Errorf("Expected name [%v], got [%v]", expectedName, b.Name)
			}
			if b.Age != expectedAge {
				t.Errorf("Expected age [%v], got [%v]", expectedAge, b.Age)
			}
		}
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	c := NewHttpClient("POST", server.URL)
	c.AddJsonBody(bodyRequest{
		Name: expectedName,
		Age:  expectedAge,
	})
	_, err := c.Send()
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func TestHttpClientWithXmlBodyRequest(t *testing.T) {
	type bodyRequest struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}
	expectedName := "some_value"
	expectedAge := 23

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		expectedContentType := "application/xml"
		contentType := req.Header.Get("Content-Type")
		if expectedContentType != contentType {
			t.Errorf("Expected Content-Type [%v], got [%v]", expectedContentType, contentType)
		}

		body, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			t.Errorf("Error occurred when read body [%v]", err)
		} else {
			b := bodyRequest{}
			err = xml.Unmarshal(body, &b)
			if err != nil {
				t.Errorf("Error occurred when unmarshal xml [%v]", err)
			}
			if b.Name != expectedName {
				t.Errorf("Expected name [%v], got [%v]", expectedName, b.Name)
			}
			if b.Age != expectedAge {
				t.Errorf("Expected age [%v], got [%v]", expectedAge, b.Age)
			}
		}
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	c := NewHttpClient("POST", server.URL)
	c.AddXmlBody(bodyRequest{
		Name: expectedName,
		Age:  expectedAge,
	})
	_, err := c.Send()
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func TestHttpClientWithUnmarshalJsonBodyResponse(t *testing.T) {
	type bodyRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	expectedName := "some_value"
	expectedAge := 23

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		res, err := json.Marshal(bodyRequest{
			Name: expectedName,
			Age:  expectedAge,
		})
		if err != nil {
			t.Errorf("Error occurred [%v]", err)
		}
		w.Write(res)
	}))
	defer server.Close()

	c := NewHttpClient("GET", server.URL)
	c.Send()

	b := bodyRequest{}
	err := c.UnmarshalJson(&b)
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
	if b.Name != expectedName {
		t.Errorf("Expected name [%v], got [%v]", expectedName, b.Name)
	}
	if b.Age != expectedAge {
		t.Errorf("Expected age [%v], got [%v]", expectedAge, b.Age)
	}
}

func TestHttpClientWithUnmarshalXmlBodyResponse(t *testing.T) {
	type bodyRequest struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}
	expectedName := "some_value"
	expectedAge := 23

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		res, err := xml.Marshal(bodyRequest{
			Name: expectedName,
			Age:  expectedAge,
		})
		if err != nil {
			t.Errorf("Error occurred [%v]", err)
		}
		w.Write(res)
	}))
	defer server.Close()

	c := NewHttpClient("GET", server.URL)
	c.Send()

	b := bodyRequest{}
	err := c.UnmarshalXml(&b)
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
	if b.Name != expectedName {
		t.Errorf("Expected name [%v], got [%v]", expectedName, b.Name)
	}
	if b.Age != expectedAge {
		t.Errorf("Expected age [%v], got [%v]", expectedAge, b.Age)
	}
}

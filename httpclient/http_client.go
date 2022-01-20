package grest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	"grest.dev/grest/convert"
)

type HttpClient struct {
	IsDebug      bool
	Method       string
	Url          string
	Header       http.Header
	Body         io.Reader
	BodyDebug    []byte
	BodyResponse []byte
}

func New(method, url string) *HttpClient {
	return &HttpClient{Method: method, Url: url}
}

func (c *HttpClient) AddHeader(key, value string) {
	if c.Header == nil {
		c.Header = http.Header{key: []string{value}}
	} else if len(c.Header[key]) == 0 {
		c.Header[key] = []string{value}
	} else {
		c.Header[key] = append(c.Header[key], value)
	}
}

func (c *HttpClient) AddMultipartBody(body interface{}) error {
	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)
	data, ok := body.(map[string]interface{})
	if ok {
		for k, v := range data {
			f, ok := v.(*os.File)
			if ok {
				part, err := writer.CreateFormFile(k, f.Name())
				if err != nil {
					return err
				}
				_, errCopy := io.Copy(part, f)
				if errCopy != nil {
					return errCopy
				}
			} else {
				writer.WriteField(k, fmt.Sprintf("%v", v))
			}
		}
	} else {
		v := reflect.ValueOf(body)
		t := reflect.TypeOf(body)
		if t.Kind() == reflect.Ptr {
			v = reflect.ValueOf(body).Elem()
			t = v.Type()
		}
		if t.Kind() == reflect.Struct {
			for i := 0; i < t.NumField(); i++ {
				key := t.Field(i).Tag.Get("form")
				if key == "" {
					key = t.Field(i).Tag.Get("json")
				}
				if key == "" {
					key = convert.ToSnakeCase(t.Field(i).Name)
				}
				val := v.Field(i).Interface()
				f, ok := val.(*os.File)
				if ok {
					part, err := writer.CreateFormFile(key, f.Name())
					if err != nil {
						return err
					}
					_, errCopy := io.Copy(part, f)
					if errCopy != nil {
						return errCopy
					}
				} else {
					writer.WriteField(key, fmt.Sprintf("%v", val))
				}
			}
		}
	}
	err := writer.Close()
	if err != nil {
		return err
	}
	c.Body = b
	if c.IsDebug {
		c.BodyDebug, _ = json.MarshalIndent(body, "", "  ")
	}
	c.AddHeader("Content-Type", writer.FormDataContentType())
	return nil
}

func (c *HttpClient) AddUrlEncodedBody(body interface{}) error {
	params := url.Values{}
	data, ok := body.(map[string]interface{})
	if ok {
		for k, v := range data {
			params.Add(k, fmt.Sprintf("%v", v))
		}
	} else {
		v := reflect.ValueOf(body)
		t := reflect.TypeOf(body)
		if t.Kind() == reflect.Ptr {
			v = reflect.ValueOf(body).Elem()
			t = v.Type()
		}
		if t.Kind() == reflect.Struct {
			for i := 0; i < t.NumField(); i++ {
				key := t.Field(i).Tag.Get("form")
				if key == "" {
					key = t.Field(i).Tag.Get("json")
				}
				if key == "" {
					key = convert.ToSnakeCase(t.Field(i).Name)
				}
				val := v.Field(i).Interface()
				params.Add(key, fmt.Sprintf("%v", val))
			}
		}
	}

	b := []byte(params.Encode())
	c.Body = bytes.NewBuffer(b)
	if c.IsDebug {
		c.BodyDebug = b
	}
	c.AddHeader("Content-Type", "application/x-www-form-urlencoded")
	return nil
}

func (c *HttpClient) AddJsonBody(body interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	c.Body = bytes.NewBuffer(b)
	if c.IsDebug {
		c.BodyDebug, _ = json.MarshalIndent(body, "", "  ")
	}
	c.AddHeader("Content-Type", "application/json")
	return nil
}

func (c *HttpClient) AddXmlBody(body interface{}) error {
	b, err := xml.Marshal(body)
	if err != nil {
		return err
	}
	c.Body = bytes.NewBuffer(b)
	if c.IsDebug {
		c.BodyDebug, _ = xml.MarshalIndent(body, "", "  ")
	}
	c.AddHeader("Content-Type", "application/xml")
	return nil
}

func (c *HttpClient) Send() (*http.Response, error) {
	req, err := http.NewRequest(c.Method, c.Url, c.Body)
	if err != nil {
		return nil, err
	}
	for k, v := range c.Header {
		for _, h := range v {
			req.Header.Add(k, h)
		}
	}

	if c.IsDebug {
		fmt.Println("-----------------------------------------------------")
		fmt.Println(c.Method, c.Url)
		fmt.Println(string(c.BodyDebug))
	}

	startTime := time.Now()
	res, err := http.DefaultClient.Do(req)
	endTime := time.Now()
	responseTime := endTime.Sub(startTime).Seconds()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	c.BodyResponse = b
	if c.IsDebug {
		fmt.Println("-----------------------------------------------------")
		fmt.Println("http status code:", res.StatusCode, ", response time:", fmt.Sprintf("%v", responseTime)+"s")
		fmt.Println(string(b))
	}

	if res.StatusCode >= 400 && res.StatusCode <= 499 {
		return res, errors.New("Client Error")
	} else if res.StatusCode >= 500 && res.StatusCode <= 599 {
		return res, errors.New("Server Error")
	} else if res.StatusCode < 100 || res.StatusCode >= 600 {
		return res, errors.New("Unknown Error")
	}

	return res, nil
}

func (c *HttpClient) UnmarshalJson(v interface{}) error {
	return json.Unmarshal(c.BodyResponse, v)
}

func (c *HttpClient) UnmarshalXml(v interface{}) error {
	return xml.Unmarshal(c.BodyResponse, v)
}

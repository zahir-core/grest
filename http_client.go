package grest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"
)

type HttpClient struct {
	IsDebug      bool
	Timeout      time.Duration
	Method       string
	Url          string
	Header       http.Header
	Body         io.Reader
	BodyDebug    []byte
	BodyResponse []byte
}

func NewHttpClient(method, url string) *HttpClient {
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

func (c *HttpClient) AddMultipartBody(body any) error {
	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)
	data, ok := body.(map[string]any)
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
					key = String{}.SnakeCase(t.Field(i).Name)
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

func (c *HttpClient) AddUrlEncodedBody(body any) error {
	params := url.Values{}
	data, ok := body.(map[string]any)
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
					key = String{}.SnakeCase(t.Field(i).Name)
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

func (c *HttpClient) AddJsonBody(body any) error {
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

func (c *HttpClient) AddXmlBody(body any) error {
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

func (c *HttpClient) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
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
	client := http.DefaultClient
	if c.Timeout > 0 {
		client.Timeout = c.Timeout
	}
	res, err := client.Do(req)
	endTime := time.Now()
	responseTime := endTime.Sub(startTime).Seconds()
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return res, err
	}
	c.BodyResponse = b
	if c.IsDebug {
		fmt.Println("-----------------------------------------------------")
		fmt.Println("http status code:", res.StatusCode, ", response time:", fmt.Sprintf("%v", responseTime)+"s")
		fmt.Println(string(b))
	}

	if res.StatusCode >= 400 && res.StatusCode <= 499 {
		return res, NewError(res.StatusCode, "Client Error")
	} else if res.StatusCode >= 500 && res.StatusCode <= 599 {
		return res, NewError(res.StatusCode, "Server Error")
	} else if res.StatusCode < 100 || res.StatusCode >= 400 {
		return res, NewError(res.StatusCode, "Unknown Error")
	}

	return res, nil
}

func (c *HttpClient) BodyResponseStr() string {
	return string(c.BodyResponse)
}

func (c *HttpClient) UnmarshalJson(v any) error {
	return json.Unmarshal(c.BodyResponse, v)
}

func (c *HttpClient) UnmarshalXml(v any) error {
	return xml.Unmarshal(c.BodyResponse, v)
}

func (c *HttpClient) Debug() {
	c.IsDebug = true
}

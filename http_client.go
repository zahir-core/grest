package grest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"
)

// HttpClient is a utility to perform HTTP requests and manage various request parameters.
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

// NewHttpClient creates a new HttpClient instance with the provided HTTP method and URL.
func NewHttpClient(method, url string) *HttpClient {
	return &HttpClient{Method: method, Url: url}
}

// AddHeader adds a new header to the request.
func (c *HttpClient) AddHeader(key, value string) {
	if c.Header == nil {
		c.Header = http.Header{key: []string{value}}
	} else if len(c.Header[key]) == 0 {
		c.Header[key] = []string{value}
	} else {
		c.Header[key] = append(c.Header[key], value)
	}
}

// AddMultipartBody adds a multipart/form-data request body.
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

// AddUrlEncodedBody adds an application/x-www-form-urlencoded request body.
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

// AddJsonBody adds a JSON request body.
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

// AddXmlBody adds an XML request body.
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

// SetTimeout sets the timeout duration for the HTTP request.
func (c *HttpClient) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// Send sends the HTTP request and returns the HTTP response.
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
	defaultTransport := &http.Transport{}
	defaultTransport.DialContext = (&net.Dialer{}).DialContext
	client := http.Client{Timeout: time.Second * 10, Transport: defaultTransport}
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

// BodyResponseStr returns the response body as a string.
func (c *HttpClient) BodyResponseStr() string {
	return string(c.BodyResponse)
}

// UnmarshalJson unmarshals the JSON response body into the provided target.
func (c *HttpClient) UnmarshalJson(v any) error {
	return json.Unmarshal(c.BodyResponse, v)
}

// UnmarshalXml unmarshals the XML response body into the provided target.
func (c *HttpClient) UnmarshalXml(v any) error {
	return xml.Unmarshal(c.BodyResponse, v)
}

// Debug enables debug mode for the HttpClient, printing request and response details.
func (c *HttpClient) Debug() {
	c.IsDebug = true
}

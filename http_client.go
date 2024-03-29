package grest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"
)

// HttpClient is a utility to perform HTTP requests and manage various request parameters.
type HttpClient struct {
	IsDebug      bool
	Logger       LoggerInterface
	Client       *http.Client
	Method       string
	Url          string
	Header       http.Header
	Body         io.Reader
	BodyRequest  []byte
	BodyResponse []byte
}

// NewHttpClient creates a new HttpClient instance with the provided HTTP method and URL.
func NewHttpClient(method, url string) *HttpClient {
	return &HttpClient{Method: method, Url: url, Client: http.DefaultClient}
}

// SetClient set your own http client instead of http.DefaultClient
func (c *HttpClient) SetClient(client *http.Client) {
	c.Client = client
}

// SetTimeout sets the timeout duration for the HTTP request.
func (c *HttpClient) SetTimeout(timeout time.Duration) {
	c.Client.Timeout = timeout
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
		c.BodyRequest, _ = json.Marshal(body)
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
	c.BodyRequest = b
	c.Body = bytes.NewBuffer(b)
	c.AddHeader("Content-Type", "application/x-www-form-urlencoded")
	return nil
}

// AddJsonBody adds a JSON request body.
func (c *HttpClient) AddJsonBody(body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	c.BodyRequest = b
	c.Body = bytes.NewBuffer(b)
	c.AddHeader("Content-Type", "application/json")
	return nil
}

// AddXmlBody adds an XML request body.
func (c *HttpClient) AddXmlBody(body any) error {
	b, err := xml.Marshal(body)
	if err != nil {
		return err
	}
	c.BodyRequest = b
	c.Body = bytes.NewBuffer(b)
	c.AddHeader("Content-Type", "application/xml")
	return nil
}

// Send sends the HTTP request and returns the HTTP response.
func (c *HttpClient) Send() (*http.Response, error) {
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
	logger := c.Logger.Info
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
		c.Logger.Debug("HttpClient Request",
			slog.String("method", c.Method),
			slog.String("url", c.Url),
			slog.Any("header", c.Header),
			slog.String("body", string(c.BodyRequest)),
		)
	}

	startTime := time.Now()
	res, err := c.Client.Do(req)
	if err == nil {
		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		if err == nil {
			c.BodyResponse = b
		}
	}
	if res == nil {
		res = &http.Response{
			StatusCode: 500,
			Request:    req,
		}
	}
	logAttrs := []any{}
	if c.IsDebug {
		logAttrs = []any{
			slog.Any("error", err),
			slog.Group("request",
				slog.String("method", c.Method),
				slog.String("url", c.Url),
				slog.Any("header", c.Header),
				slog.String("body", string(c.BodyRequest)),
				slog.Time("at", startTime),
			),
			slog.Group("response",
				slog.Duration("duration", time.Now().Sub(startTime)),
				slog.Int("statusCode", res.StatusCode),
				slog.String("status", res.Status),
				slog.Any("header", res.Header),
				slog.String("body", string(c.BodyResponse)),
			),
		}
	}

	if res.StatusCode >= 400 && res.StatusCode <= 499 {
		logger = c.Logger.Warn
		err = NewError(res.StatusCode, "Client Error")
	} else if res.StatusCode >= 500 && res.StatusCode <= 599 {
		logger = c.Logger.Error
		err = NewError(res.StatusCode, "Server Error")
	} else if res.StatusCode < 100 || res.StatusCode >= 400 {
		logger = c.Logger.Error
		err = NewError(res.StatusCode, "Unknown Error")
	}
	if c.IsDebug {
		logger("HttpClient Response", logAttrs...)
	}

	return res, err
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

// SetLogger set logger
func (c *HttpClient) SetLogger(logger LoggerInterface) {
	c.Logger = logger
}

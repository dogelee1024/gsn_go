package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/openwallet1/gsn-go/common/errs"
	"github.com/openwallet1/gsn-go/common/utils"
)

type HttpCli struct {
	httpClient   *http.Client
	httpRequest  *http.Request
	httpResponse *http.Response
	Error        error
}

func newHttpClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func GetWithHeader[T any](url string, params any, headers map[string]string) (*T, error) {
	var resp ApiResponse
	reqURL := url
	queryParams, err := structToURLParams(params)
	if err != nil {
		return nil, err
	}

	// 构建完整的URL
	if queryParams != "" {
		reqURL = fmt.Sprintf("%s?%s", url, queryParams)
	}

	client := Get(reqURL)
	for k, v := range headers {
		client.SetHeader(k, v)
	}

	err = client.ToJson(&resp)
	if err != nil {
		return nil, err
	}

	fmt.Println("get resp from remote server", utils.StructToJsonString(resp))

	if len(resp.Data) == 0 || resp.Data == nil {
		return nil, errs.ErrRecordNotFound
	}

	var specificData T
	if err := json.Unmarshal(resp.Data, &specificData); err != nil {
		return nil, errs.Wrap(err, "Unmarshal Data failed, url")

	}
	return &specificData, nil
}

func PostToRelay[T any](url string, data interface{}, headers map[string]string) (*T, error) {
	client := Post(url).BodyWithJson(data)
	var specificData T

	for k, v := range headers {
		client.SetHeader(k, v)
	}
	err := client.ToJson2(&specificData)
	if err != nil {
		return nil, err
	}
	return &specificData, nil
}

func PostWithHeader[T any](url string, data interface{}, headers map[string]string) (*T, error) {
	var resp ApiResponse
	client := Post(url).BodyWithJson(data)

	for k, v := range headers {
		client.SetHeader(k, v)
	}
	err := client.ToJson(&resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 || resp.Data == nil {
		return nil, errs.ErrRecordNotFound
	}
	var specificData T
	if err := json.Unmarshal(resp.Data, &specificData); err != nil {
		return nil, Wrap(err, "Unmarshal Data failed, url")
	}

	cookie := client.httpResponse.Header.Get("Set-Cookie")
	if cookie != "" {
		headers["Cookie"] = cookie
	}
	return &specificData, nil
}

func Get(url string) *HttpCli {
	request, err := http.NewRequest("GET", url, nil)
	return &HttpCli{
		httpClient:  newHttpClient(),
		httpRequest: request,
		Error:       err,
	}
}

func Post(url string) *HttpCli {
	request, err := http.NewRequest("POST", url, nil)
	return &HttpCli{
		httpClient:  newHttpClient(),
		httpRequest: request,
		Error:       Wrap(err, "newRequest failed, url"),
	}
}

func (c *HttpCli) SetTimeOut(timeout time.Duration) *HttpCli {
	c.httpClient.Timeout = timeout
	return c
}

func (c *HttpCli) SetHeader(key, value string) *HttpCli {
	c.httpRequest.Header.Set(key, value)
	return c
}

func structToURLParams(params any) (string, error) {
	if params == nil || params == "" {
		return "", nil
	}
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	var queryParams []string
	err = appendFieldsValues(jsonData, &queryParams, "")
	if err != nil {
		return "", err
	}

	return strings.Join(queryParams, "&"), nil
}

func appendFieldsValues(v any, queryParams *[]string, prefix string) error {
	var data map[string]interface{}
	if bytes, ok := v.([]byte); ok {
		err := json.Unmarshal(bytes, &data)
		if err != nil {
			return errs.ErrHttpArgs
		} else {
			for key, value := range data {
				if nestedData, ok := value.(map[string]interface{}); ok {
					nestedPrefix := key
					if prefix != "" {
						nestedPrefix = prefix + "." + key
					}
					err := appendFieldsValues(nestedData, queryParams, nestedPrefix)
					if err != nil {
						return err
					}
				} else {
					var fieldName string
					if prefix != "" {
						fieldName = prefix + "." + key
					} else {
						fieldName = key
					}
					*queryParams = append(*queryParams, fmt.Sprintf("%s=%v", url.QueryEscape(fieldName), url.QueryEscape(fmt.Sprintf("%v", value))))
				}
			}
		}
	} else {
		return errs.ErrHttpArgs
	}
	return nil
}

func (c *HttpCli) BodyWithJson(obj interface{}) *HttpCli {
	if c.Error != nil {
		return c
	}

	buf, err := json.Marshal(obj)
	if err != nil {
		c.Error = Wrap(err, "marshal failed, url")
		return c
	}
	c.httpRequest.Body = io.NopCloser(bytes.NewReader(buf))
	c.httpRequest.ContentLength = int64(len(buf))
	c.httpRequest.Header.Set("Content-Type", "application/json")
	return c
}

func (c *HttpCli) BodyWithBytes(buf []byte) *HttpCli {
	if c.Error != nil {
		return c
	}

	c.httpRequest.Body = io.NopCloser(bytes.NewReader(buf))
	c.httpRequest.ContentLength = int64(len(buf))
	return c
}

func (c *HttpCli) BodyWithForm(form map[string]string) *HttpCli {
	if c.Error != nil {
		return c
	}

	var value url.Values = make(map[string][]string, len(form))
	for k, v := range form {
		value.Add(k, v)
	}
	buf := Str2bytes(value.Encode())

	c.httpRequest.Body = io.NopCloser(bytes.NewReader(buf))
	c.httpRequest.ContentLength = int64(len(buf))
	c.httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func (c *HttpCli) ToBytes() (content []byte, err error) {
	if c.Error != nil {
		return nil, c.Error
	}

	resp, err := c.httpClient.Do(c.httpRequest)
	if err != nil {
		return nil, Wrap(err, "client.Do failed, url")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, Wrap(errors.New(resp.Status), "status code failed ")
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Wrap(err, "io.ReadAll failed, url")
	}

	return buf, nil
}

func (c *HttpCli) ToJson2(obj interface{}) error {
	if c.Error != nil {
		return c.Error
	}

	resp, err := c.httpClient.Do(c.httpRequest)
	if err != nil {
		return Wrap(err, "client.Do failed, url")
	}
	c.httpResponse = resp
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Wrap(errors.New(resp.Status), "status code failed ")
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return Wrap(err, "io.ReadAll failed, url")
	}
	err = json.Unmarshal(buf, obj)
	if err != nil {
		return Wrap(err, "marshal failed, url")
	}

	return nil
}

func (c *HttpCli) ToJson(obj *ApiResponse) error {
	if c.Error != nil {
		return c.Error
	}

	resp, err := c.httpClient.Do(c.httpRequest)
	if err != nil {
		return Wrap(err, "client.Do failed, url")
	}
	c.httpResponse = resp
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Wrap(errors.New(resp.Status), "status code failed ")
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return Wrap(err, "io.ReadAll failed, url")
	}
	err = json.Unmarshal(buf, obj)
	if err != nil {
		return Wrap(err, "marshal failed, url")
	}

	return nil
}

func Wrap(err error, message string) error {
	return errs.Wrap(err, "==> "+printCallerNameAndLine()+message)
}

func printCallerNameAndLine() string {
	pc, _, line, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name() + "()@" + strconv.Itoa(line) + ": "
}

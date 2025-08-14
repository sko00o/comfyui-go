package comfyui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	comfyError "github.com/sko00o/comfyui-go/error"
	"github.com/sko00o/comfyui-go/logger"
)

type Config struct {
	Endpoint string `mapstructure:"endpoint"`

	Timeout       time.Duration `mapstructure:"timeout"`
	DialTimeout   time.Duration `mapstructure:"dial_timeout"`
	DialKeepAlive time.Duration `mapstructure:"dial_keep_alive"`
}

type Client struct {
	*http.Client
	BaseURL url.URL
	apiMap  sync.Map

	log logger.LoggerExtend
}

type Option func(c *Client)

func WithLogger(l logger.LoggerExtend) Option {
	return func(c *Client) {
		c.log = l
	}
}

func New(c Config, opts ...Option) (*Client, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	cli := &Client{
		Client: &http.Client{
			Timeout: c.Timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   c.DialTimeout,
					KeepAlive: c.DialKeepAlive,
				}).DialContext,
			},
		},
		BaseURL: *u,

		log: logger.NewStd(),
	}
	for _, opt := range opts {
		opt(cli)
	}

	return cli, nil
}

type api struct {
	Once sync.Once
	URL  string
}

func (c *Client) reqURL(path ReqPath) string {
	got, _ := c.apiMap.LoadOrStore(path, &api{})
	a := got.(*api)
	a.Once.Do(func() {
		u := c.BaseURL
		u.Path = string(path)
		a.URL = u.String()
	})
	return a.URL
}

func (c *Client) newJSONReq(method, urlStr string, data any) (req *http.Request, err error) {
	var body io.Reader
	if data != nil {
		var buf bytes.Buffer
		if err = json.NewEncoder(&buf).Encode(data); err != nil {
			return nil, fmt.Errorf("encode: %w", err)
		}
		body = &buf
	}

	req, err = http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return
}

func (c *Client) reqJSON(method, urlStr string, data any) (*http.Response, error) {
	req, err := c.newJSONReq(method, urlStr, data)
	if err != nil {
		return nil, fmt.Errorf("newJSONReq: %w", err)
	}
	return c.Do(req)
}

type getRespFunc func() (*http.Response, error)

func (c *Client) postJSON(path ReqPath, data any) getRespFunc {
	return func() (*http.Response, error) {
		return c.reqJSON(http.MethodPost, c.reqURL(path), data)
	}
}

func (c *Client) getJSON(path ReqPath, values url.Values) getRespFunc {
	return func() (*http.Response, error) {
		urlStr := c.reqURL(path)
		if len(values) != 0 {
			urlStr += "?" + values.Encode()
		}
		return c.reqJSON(http.MethodGet, urlStr, nil)
	}
}

type handleRespFunc func(rd io.Reader, header http.Header) error

func (c *Client) process(run getRespFunc, handle handleRespFunc) error {
	resp, err := run()
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		errMsg, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
		// empty body, use status as error
		if len(errMsg) == 0 {
			return fmt.Errorf("comfyui status: %s", resp.Status)
		}
		return comfyError.ComfyUIError{Message: json.RawMessage(errMsg)}
	}

	if handle == nil {
		return nil
	}
	return handle(resp.Body, resp.Header)
}

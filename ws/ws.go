package ws

import (
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/sko00o/comfyui-go/logger"
)

type Client struct {
	ClientID string
	urlStr   string
	handler  Handler
	log      logger.Logger

	done      chan struct{}
	isClosing atomic.Bool

	conn *websocket.Conn
}

func New(u url.URL, clientID string, handler Handler, l logger.LoggerExtend) (*Client, error) {
	u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1)
	u.Path = "/ws"
	q := u.Query()
	q.Set("clientId", clientID)
	u.RawQuery = q.Encode()
	urlStr := u.String()

	c := &Client{
		ClientID: clientID,
		urlStr:   urlStr,
		handler:  handler,
		log:      l.With("client_id", clientID),

		done:      make(chan struct{}),
		isClosing: atomic.Bool{},
	}

	return c, c.connect()
}

func (c *Client) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.urlStr, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	c.log.Debugf("ws connected")
	go c.readLoop()
	return nil
}

func (c *Client) reconnect() {
	c.log.Warnf("ws reconnecting...")
	if c.conn != nil {
		_ = c.conn.Close()
	}
	for {
		if err := c.connect(); err != nil {
			c.log.Errorf("ws reconnect: %v", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}

	c.isClosing.Store(true)
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	select {
	case <-c.done:
	case <-time.After(5 * time.Second):
	}
	return c.conn.Close()
}

func (c *Client) readLoop() {
	defer func() {
		if c.isClosing.Load() {
			close(c.done)
			c.log.Debugf("ws closing...")
		} else {
			c.reconnect()
		}
	}()
	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) && c.isClosing.Load() {
				return
			}
			c.log.Errorf("ws read: %v", err)
			return
		}

		if c.handler != nil {
			c.handler.HandleMessage(messageType, message)
		}
	}
}

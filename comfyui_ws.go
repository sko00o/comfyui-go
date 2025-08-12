package comfyui

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"sync"

	"github.com/sko00o/comfyui-go/iface"
	"github.com/sko00o/comfyui-go/logger"
	"github.com/sko00o/comfyui-go/ws"
)

type SimpleWsClient struct {
	upstream *ws.Client
	conn     iface.MessageHandler
	log      logger.Logger
}

func NewSimpleWsClient(BaseURL url.URL, clientID string, consumer iface.MessageHandler, log logger.LoggerExtend) (*SimpleWsClient, error) {
	client := &SimpleWsClient{
		log:  log,
		conn: consumer,
	}
	upstream, err := ws.New(
		BaseURL,
		clientID,
		ws.HandlerFunc(func(messageType int, payload []byte) {
			if err := consumer.WriteMessage(messageType, payload); err != nil {
				log.Errorf("write message to client failed: %v", err)
			}
		}),
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("create upstream client failed: %w", err)
	}
	client.upstream = upstream

	return client, nil
}

func (c *SimpleWsClient) Close() error {
	if err := c.conn.Close(); err != nil {
		c.log.Errorf("close consumer failed: %v", err)
	}
	return c.upstream.Close()
}

func (c *Client) SimpleProcess(id string, consumer iface.MessageHandler) (*sync.WaitGroup, error) {
	client, err := NewSimpleWsClient(c.BaseURL, id, consumer, c.log)
	if err != nil {
		return nil, fmt.Errorf("new wsClient: %w", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer client.Close()
		for {
			_, _, err := consumer.ReadMessage()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					c.log.With("client_id", id).Warnf("ws consumer %s read: %v", consumer.Name(), err)
				}
				break
			}
		}
	}()
	return wg, nil
}

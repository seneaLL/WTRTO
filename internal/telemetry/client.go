package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"
	"time"
)

const DefaultBaseURL = "http://localhost:8111"

var ErrGameNotRunning = errors.New("telemetry: war thunder local api is unreachable (game not running?)")

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 2 * time.Second},
	}
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return ErrGameNotRunning
		}

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telemetry: unexpected status %s from %s", resp.Status, path)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, out)
}

func (c *Client) Indicators(ctx context.Context) (*Indicators, error) {
	var v Indicators
	if err := c.get(ctx, "/indicators", &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Client) State(ctx context.Context) (State, error) {
	var v State
	if err := c.get(ctx, "/state", &v); err != nil {
		return nil, err
	}

	return v, nil
}

func (c *Client) MapInfo(ctx context.Context) (*MapInfo, error) {
	var v MapInfo
	if err := c.get(ctx, "/map_info.json", &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Client) MapObjects(ctx context.Context) ([]MapObject, error) {
	var v []MapObject
	if err := c.get(ctx, "/map_obj.json", &v); err != nil {
		return nil, err
	}

	return v, nil
}

func (c *Client) HudMsg(ctx context.Context) (*HudMsg, error) {
	var v HudMsg
	if err := c.get(ctx, "/hudmsg", &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Client) Mission(ctx context.Context) (*Mission, error) {
	var v Mission
	if err := c.get(ctx, "/mission.json", &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Client) GameChat(ctx context.Context) ([]ChatMessage, error) {
	var v []ChatMessage
	if err := c.get(ctx, "/gamechat", &v); err != nil {
		return nil, err
	}

	return v, nil
}

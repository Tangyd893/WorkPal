package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type IMClient struct {
	baseURL       string
	internalToken string
	client        *http.Client
}

type apiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func NewIMClient(baseURL string, internalToken ...string) *IMClient {
	token := ""
	if len(internalToken) > 0 {
		token = internalToken[0]
	}
	return &IMClient{
		baseURL:       strings.TrimRight(baseURL, "/"),
		internalToken: token,
		client:        &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *IMClient) IsMember(ctx context.Context, convID, userID int64) (bool, error) {
	var body apiResponse[struct {
		IsMember bool `json:"is_member"`
	}]
	if err := c.get(ctx, fmt.Sprintf("/internal/conversations/%d/members/%d", convID, userID), &body); err != nil {
		return false, err
	}
	if body.Code != 0 {
		return false, fmt.Errorf("im service rejected member check: %s", body.Message)
	}
	return body.Data.IsMember, nil
}

func (c *IMClient) ListUserConversationIDs(ctx context.Context, userID int64) ([]int64, error) {
	var body apiResponse[struct {
		ConvIDs []int64 `json:"conv_ids"`
	}]
	if err := c.get(ctx, fmt.Sprintf("/internal/users/%d/conversations", userID), &body); err != nil {
		return nil, err
	}
	if body.Code != 0 {
		return nil, fmt.Errorf("im service rejected conversation list: %s", body.Message)
	}
	return body.Data.ConvIDs, nil
}

func (c *IMClient) get(ctx context.Context, path string, out interface{}) error {
	target, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	if c.internalToken != "" {
		req.Header.Set("X-Internal-Token", c.internalToken)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("im service returned HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

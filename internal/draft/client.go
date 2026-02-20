package draft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"wxcli/internal/errfmt"
	"wxcli/internal/httpclient"
)

const defaultBaseURL = "https://api.weixin.qq.com"

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

type DraftArticle struct {
	ArticleType        string       `json:"article_type,omitempty"`
	Title              string       `json:"title,omitempty"`
	Author             string       `json:"author,omitempty"`
	Digest             string       `json:"digest,omitempty"`
	Content            string       `json:"content,omitempty"`
	ContentSourceURL   string       `json:"content_source_url,omitempty"`
	ThumbMediaID       string       `json:"thumb_media_id,omitempty"`
	NeedOpenComment    int          `json:"need_open_comment,omitempty"`
	OnlyFansCanComment int          `json:"only_fans_can_comment,omitempty"`
	URL                string       `json:"url,omitempty"`
	ImageInfo          *ImageInfo   `json:"image_info,omitempty"`
	ProductInfo        *ProductInfo `json:"product_info,omitempty"`
}

type ImageInfo struct {
	ImageList []ImageItem `json:"image_list,omitempty"`
}

type ImageItem struct {
	ImageMediaID string `json:"image_media_id,omitempty"`
}

type ProductInfo struct {
	FooterProductInfo ProductFooter `json:"footer_product_info,omitempty"`
}

type ProductFooter struct {
	ProductKey string `json:"product_key,omitempty"`
}

type AddDraftRequest struct {
	Articles []DraftArticle `json:"articles"`
}

type AddDraftResponse struct {
	MediaID string `json:"media_id"`
}

type GetDraftRequest struct {
	MediaID string `json:"media_id"`
}

type GetDraftResponse struct {
	NewsItem []DraftArticle `json:"news_item"`
}

type BatchGetRequest struct {
	Offset    int `json:"offset"`
	Count     int `json:"count"`
	NoContent int `json:"no_content,omitempty"`
}

type BatchGetResponse struct {
	TotalCount int         `json:"total_count"`
	ItemCount  int         `json:"item_count"`
	Item       []DraftItem `json:"item"`
}

type DraftItem struct {
	MediaID    string        `json:"media_id"`
	UpdateTime int64         `json:"update_time"`
	Content    *DraftContent `json:"content,omitempty"`
}

type DraftContent struct {
	NewsItem []DraftArticle `json:"news_item"`
}

type DeleteDraftRequest struct {
	MediaID string `json:"media_id"`
}

type DeleteDraftResponse struct{}

type apiErrorEnvelope struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Transport: httpclient.NewRetryTransport(nil)}
	}
	return &Client{BaseURL: defaultBaseURL, HTTP: httpClient}
}

func (c *Client) Add(ctx context.Context, accessToken string, req AddDraftRequest) (AddDraftResponse, error) {
	var resp AddDraftResponse
	err := c.post(ctx, accessToken, "/cgi-bin/draft/add", req, &resp)
	return resp, err
}

func (c *Client) Get(ctx context.Context, accessToken string, req GetDraftRequest) (GetDraftResponse, error) {
	var resp GetDraftResponse
	err := c.post(ctx, accessToken, "/cgi-bin/draft/get", req, &resp)
	return resp, err
}

func (c *Client) List(ctx context.Context, accessToken string, req BatchGetRequest) (BatchGetResponse, error) {
	var resp BatchGetResponse
	err := c.post(ctx, accessToken, "/cgi-bin/draft/batchget", req, &resp)
	return resp, err
}

func (c *Client) Delete(ctx context.Context, accessToken string, req DeleteDraftRequest) error {
	var resp DeleteDraftResponse
	return c.post(ctx, accessToken, "/cgi-bin/draft/delete", req, &resp)
}

func (c *Client) post(ctx context.Context, accessToken, path string, body any, out any) error {
	if accessToken == "" {
		return fmt.Errorf("missing access_token")
	}
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	if !strings.HasPrefix(base, "http") {
		base = "https://" + base
	}
	endpoint, err := url.JoinPath(base, path)
	if err != nil {
		return err
	}
	endpoint += "?access_token=" + url.QueryEscape(accessToken)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(body); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env apiErrorEnvelope
	_ = json.Unmarshal(data, &env)
	if env.ErrCode != 0 {
		return &errfmt.APIError{Code: env.ErrCode, Message: env.ErrMsg}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}

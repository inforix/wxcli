package material

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"wxcli/src/internal/errfmt"
	"wxcli/src/internal/httpclient"
)

const defaultBaseURL = "https://api.weixin.qq.com"

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

type GetMaterialRequest struct {
	MediaID string `json:"media_id"`
}

type MaterialNewsItem struct {
	Title              string `json:"title,omitempty"`
	Author             string `json:"author,omitempty"`
	Digest             string `json:"digest,omitempty"`
	Content            string `json:"content,omitempty"`
	ContentSourceURL   string `json:"content_source_url,omitempty"`
	ThumbMediaID       string `json:"thumb_media_id,omitempty"`
	ShowCoverPic       int    `json:"show_cover_pic,omitempty"`
	URL                string `json:"url,omitempty"`
	ThumbURL           string `json:"thumb_url,omitempty"`
	NeedOpenComment    int    `json:"need_open_comment,omitempty"`
	OnlyFansCanComment int    `json:"only_fans_can_comment,omitempty"`
}

type NewsMaterial struct {
	NewsItem []MaterialNewsItem `json:"news_item"`
}

type VideoMaterial struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DownURL     string `json:"down_url"`
}

type UploadMaterialRequest struct {
	Type             string
	FilePath         string
	VideoTitle       string
	VideoDescription string
}

type UploadMaterialResponse struct {
	MediaID string `json:"media_id"`
	URL     string `json:"url,omitempty"`
}

type UploadImageResponse struct {
	URL string `json:"url"`
}

type NewsArticle struct {
	Title              string `json:"title,omitempty"`
	ThumbMediaID       string `json:"thumb_media_id,omitempty"`
	Author             string `json:"author,omitempty"`
	Digest             string `json:"digest,omitempty"`
	ShowCoverPic       int    `json:"show_cover_pic,omitempty"`
	Content            string `json:"content,omitempty"`
	ContentSourceURL   string `json:"content_source_url,omitempty"`
	NeedOpenComment    int    `json:"need_open_comment,omitempty"`
	OnlyFansCanComment int    `json:"only_fans_can_comment,omitempty"`
}

type AddNewsRequest struct {
	Articles []NewsArticle `json:"articles"`
}

type AddNewsResponse struct {
	MediaID string `json:"media_id"`
}

type UpdateNewsRequest struct {
	MediaID string      `json:"media_id"`
	Index   int         `json:"index"`
	Article NewsArticle `json:"articles"`
}

type UpdateNewsResponse struct{}

type CountResponse struct {
	VoiceCount int `json:"voice_count"`
	VideoCount int `json:"video_count"`
	ImageCount int `json:"image_count"`
	NewsCount  int `json:"news_count"`
}

type BatchGetRequest struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Count  int    `json:"count"`
}

type BatchGetResponse struct {
	TotalCount int            `json:"total_count"`
	ItemCount  int            `json:"item_count"`
	Item       []MaterialItem `json:"item"`
}

type MaterialItem struct {
	MediaID    string       `json:"media_id"`
	Name       string       `json:"name,omitempty"`
	URL        string       `json:"url,omitempty"`
	UpdateTime int64        `json:"update_time"`
	Content    *NewsContent `json:"content,omitempty"`
}

type NewsContent struct {
	NewsItem []MaterialNewsItem `json:"news_item"`
}

type GetMaterialResult struct {
	JSON        json.RawMessage
	News        *NewsMaterial
	Video       *VideoMaterial
	Data        []byte
	ContentType string
	Filename    string
}

type apiErrorEnvelope struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type rawResponse struct {
	data        []byte
	contentType string
	filename    string
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Transport: httpclient.NewRetryTransport(nil)}
	}
	return &Client{BaseURL: defaultBaseURL, HTTP: httpClient}
}

func (c *Client) Get(ctx context.Context, accessToken string, req GetMaterialRequest) (GetMaterialResult, error) {
	var result GetMaterialResult
	if req.MediaID == "" {
		return result, fmt.Errorf("missing media_id")
	}
	raw, err := c.postRaw(ctx, accessToken, "/cgi-bin/material/get_material", req)
	if err != nil {
		return result, err
	}
	if looksLikeJSON(raw.data, raw.contentType) {
		var env apiErrorEnvelope
		if err := json.Unmarshal(raw.data, &env); err == nil && env.ErrCode != 0 {
			return result, &errfmt.APIError{Code: env.ErrCode, Message: env.ErrMsg}
		}
		result.JSON = raw.data
		var news NewsMaterial
		if err := json.Unmarshal(raw.data, &news); err == nil && len(news.NewsItem) > 0 {
			result.News = &news
		}
		var video VideoMaterial
		if err := json.Unmarshal(raw.data, &video); err == nil && video.DownURL != "" {
			result.Video = &video
		}
		return result, nil
	}
	if len(raw.data) == 0 {
		return result, fmt.Errorf("empty response")
	}
	result.Data = raw.data
	result.ContentType = raw.contentType
	result.Filename = raw.filename
	return result, nil
}

func (c *Client) Count(ctx context.Context, accessToken string) (CountResponse, error) {
	var resp CountResponse
	if err := c.getJSON(ctx, accessToken, "/cgi-bin/material/get_materialcount", &resp); err != nil {
		return CountResponse{}, err
	}
	return resp, nil
}

func (c *Client) List(ctx context.Context, accessToken string, req BatchGetRequest) (BatchGetResponse, error) {
	var resp BatchGetResponse
	if err := c.postJSON(ctx, accessToken, "/cgi-bin/material/batchget_material", req, &resp); err != nil {
		return BatchGetResponse{}, err
	}
	return resp, nil
}

func (c *Client) Delete(ctx context.Context, accessToken string, mediaID string) error {
	if mediaID == "" {
		return fmt.Errorf("missing media_id")
	}
	return c.postJSON(ctx, accessToken, "/cgi-bin/material/del_material", map[string]string{"media_id": mediaID}, nil)
}

func (c *Client) Upload(ctx context.Context, accessToken string, req UploadMaterialRequest) (UploadMaterialResponse, error) {
	var resp UploadMaterialResponse
	if req.Type == "" {
		return resp, fmt.Errorf("missing type")
	}
	if req.FilePath == "" {
		return resp, fmt.Errorf("missing file")
	}
	fields := map[string]string{}
	if req.Type == "video" && (req.VideoTitle != "" || req.VideoDescription != "") {
		desc := struct {
			Title        string `json:"title,omitempty"`
			Introduction string `json:"introduction,omitempty"`
		}{
			Title:        req.VideoTitle,
			Introduction: req.VideoDescription,
		}
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(desc); err != nil {
			return resp, err
		}
		fields["description"] = strings.TrimSpace(buf.String())
	}
	query := url.Values{}
	query.Set("type", req.Type)
	data, err := c.postMultipart(ctx, accessToken, "/cgi-bin/material/add_material", query, "media", req.FilePath, fields)
	if err != nil {
		return resp, err
	}
	if err := decodeJSON(data, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) AddNews(ctx context.Context, accessToken string, req AddNewsRequest) (AddNewsResponse, error) {
	var resp AddNewsResponse
	if len(req.Articles) == 0 {
		return resp, fmt.Errorf("missing articles")
	}
	if err := c.postJSON(ctx, accessToken, "/cgi-bin/material/add_news", req, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) UpdateNews(ctx context.Context, accessToken string, req UpdateNewsRequest) error {
	if req.MediaID == "" {
		return fmt.Errorf("missing media_id")
	}
	if req.Index < 0 {
		return fmt.Errorf("invalid index")
	}
	var resp UpdateNewsResponse
	return c.postJSON(ctx, accessToken, "/cgi-bin/material/update_news", req, &resp)
}

func (c *Client) UploadImage(ctx context.Context, accessToken string, filePath string) (UploadImageResponse, error) {
	var resp UploadImageResponse
	if filePath == "" {
		return resp, fmt.Errorf("missing file")
	}
	data, err := c.postMultipart(ctx, accessToken, "/cgi-bin/media/uploadimg", nil, "media", filePath, nil)
	if err != nil {
		return resp, err
	}
	if err := decodeJSON(data, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) postRaw(ctx context.Context, accessToken, apiPath string, body any) (rawResponse, error) {
	var resp rawResponse
	if accessToken == "" {
		return resp, fmt.Errorf("missing access_token")
	}
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	if !strings.HasPrefix(base, "http") {
		base = "https://" + base
	}
	endpoint, err := url.JoinPath(base, apiPath)
	if err != nil {
		return resp, err
	}
	endpoint += "?access_token=" + url.QueryEscape(accessToken)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(body); err != nil {
		return resp, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return resp, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := c.HTTP.Do(req)
	if err != nil {
		return resp, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}
	resp.data = data
	resp.contentType = res.Header.Get("Content-Type")
	resp.filename = filenameFromHeader(res.Header.Get("Content-Disposition"))
	return resp, nil
}

func (c *Client) postMultipart(ctx context.Context, accessToken, apiPath string, query url.Values, fileField, filePath string, fields map[string]string) ([]byte, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("missing access_token")
	}
	if filePath == "" {
		return nil, fmt.Errorf("missing file")
	}
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	if !strings.HasPrefix(base, "http") {
		base = "https://" + base
	}
	endpoint, err := url.JoinPath(base, apiPath)
	if err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Set("access_token", accessToken)
	for key, vals := range query {
		for _, v := range vals {
			values.Add(key, v)
		}
	}
	endpoint += "?" + values.Encode()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			_ = writer.Close()
			return nil, err
		}
	}
	part, err := writer.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		_ = writer.Close()
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func (c *Client) getRaw(ctx context.Context, accessToken, apiPath string) (rawResponse, error) {
	var resp rawResponse
	if accessToken == "" {
		return resp, fmt.Errorf("missing access_token")
	}
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	if !strings.HasPrefix(base, "http") {
		base = "https://" + base
	}
	endpoint, err := url.JoinPath(base, apiPath)
	if err != nil {
		return resp, err
	}
	endpoint += "?access_token=" + url.QueryEscape(accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return resp, err
	}

	res, err := c.HTTP.Do(req)
	if err != nil {
		return resp, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}
	resp.data = data
	resp.contentType = res.Header.Get("Content-Type")
	resp.filename = filenameFromHeader(res.Header.Get("Content-Disposition"))
	return resp, nil
}

func (c *Client) postJSON(ctx context.Context, accessToken, apiPath string, body any, out any) error {
	raw, err := c.postRaw(ctx, accessToken, apiPath, body)
	if err != nil {
		return err
	}
	return decodeJSON(raw.data, out)
}

func (c *Client) getJSON(ctx context.Context, accessToken, apiPath string, out any) error {
	raw, err := c.getRaw(ctx, accessToken, apiPath)
	if err != nil {
		return err
	}
	return decodeJSON(raw.data, out)
}

func looksLikeJSON(data []byte, contentType string) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "application/json") || strings.Contains(ct, "text/plain") {
		return true
	}
	for _, b := range data {
		switch b {
		case ' ', '\n', '\r', '\t':
			continue
		case '{', '[':
			return true
		default:
			return false
		}
	}
	return false
}

func decodeJSON(data []byte, out any) error {
	var env apiErrorEnvelope
	if err := json.Unmarshal(data, &env); err == nil && env.ErrCode != 0 {
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

func filenameFromHeader(disposition string) string {
	if disposition == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return ""
	}
	if name := params["filename"]; name != "" {
		return path.Base(name)
	}
	return ""
}

package yh

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// A picture or a file is uploaded before the message that carries it is
// sent: the route hands out a Qiniu upload token, the bytes go to the
// bucket that token is for, and what comes back is the key the message
// names. The key is the MD5 of the bytes and their extension, so the same
// picture sent twice is stored once and the second send costs one request.
const (
	imageHost = "https://chat-img.jwznb.com"
	fileHost  = "https://chat-file.jwznb.com"
	audioHost = "https://chat-audio1.jwznb.com"
	videoHost = "https://chat-video1.jwznb.com"
)

// uploadKind is what a file is uploaded for: the route that hands out the
// token, the bucket it is for, where to put the bytes when the bucket
// cannot be asked, and the address a key is served from.
type uploadKind struct {
	token  string // the route that hands out the upload token
	bucket string // the bucket the token is for
	host   string // where the bytes go when the bucket cannot be asked
	serve  string // where a key is served from
}

var (
	uploadImage = uploadKind{"qiniu-token", "chat-68", "upload-z2.qiniup.com", imageHost}
	uploadFile  = uploadKind{"qiniu-token2", "chat68-file", "upload-z2.qiniup.com", fileHost}
	uploadAudio = uploadKind{"qiniu-token-audio", "chat68-file", "upload-z2.qiniup.com", audioHost}
	uploadVideo = uploadKind{"qiniu-token-video", "chat68-video", "upload-cn-east-2.qiniup.com", videoHost}
)

// UploadImage puts a picture where messages read pictures from, and returns
// the key the message names it by.
func (c *Client) UploadImage(name string, data []byte) (string, error) {
	return c.upload(uploadImage, name, data)
}

// UploadFile puts a file where messages read files from, and returns the key
// the message names it by.
func (c *Client) UploadFile(name string, data []byte) (string, error) {
	return c.upload(uploadFile, name, data)
}

// UploadVideo puts a clip in the bucket the video host serves, and returns
// the key the message names it by. A video has a bucket and a token route of
// its own: one sent with a file's token is stored where no player looks.
func (c *Client) UploadVideo(name string, data []byte) (string, error) {
	return c.upload(uploadVideo, name, data)
}

// MediaURL is the address a key is served from, as a message that names a
// key rather than an address is shown: a picture uploaded for a chat and a
// picture of a message are the same object.
func MediaURL(kind string, key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	serve := imageHost
	switch kind {
	case "file":
		serve = fileHost
	case "audio":
		serve = audioHost
	case "video":
		serve = videoHost
	}
	return serve + "/" + strings.TrimLeft(key, "/")
}

// upload puts bytes in a bucket: the token first, then the bytes, then the
// key they were stored under.
func (c *Client) upload(kind uploadKind, name string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("%w: empty upload", ErrRequest)
	}
	token, err := c.uploadToken(kind.token)
	if err != nil {
		return "", err
	}
	key := uploadKey(name, data)
	host := c.uploadHost(token, kind)
	body, contentType, err := multipartBody(token, key, name, data)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://"+host, body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrRequest, err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := c.http().Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer resp.Body.Close()
	answer, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrResponse, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrStatus, uploadReason(answer, resp.Status))
	}
	return key, nil
}

// uploadReason is what a bucket said about a file it would not take: the
// answer is a line of JSON, and its error is the part worth showing.
func uploadReason(answer []byte, status string) string {
	var e struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(answer, &e); err == nil && e.Error != "" {
		return e.Error
	}
	return status
}

// uploadToken asks for the token that lets this account put a file in a
// bucket. The route answers with JSON, and the token is in its data.
func (c *Client) uploadToken(route string) (string, error) {
	data, err := c.do("GET", "/v1/misc/"+route, nil, false)
	if err != nil {
		return "", err
	}
	var answer struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	if err := json.Unmarshal(data, &answer); err != nil {
		return "", fmt.Errorf("%w: %v", ErrResponse, err)
	}
	if answer.Code != 1 {
		return "", &APIError{Code: answer.Code, Msg: answer.Msg}
	}
	if answer.Data.Token == "" {
		return "", fmt.Errorf("%w: no upload token", ErrResponse)
	}
	return answer.Data.Token, nil
}

// uploadHost asks Qiniu where the bytes for a bucket should go. The token
// names the account, and the bucket is asked of Qiniu rather than assumed:
// a bucket lives in one region, and a request to the wrong one is refused.
// The host the client knows is used when the question cannot be asked.
func (c *Client) uploadHost(token string, kind uploadKind) string {
	ak := token
	if i := strings.Index(ak, ":"); i > 0 {
		ak = ak[:i]
	}
	if ak == "" {
		return kind.host
	}
	u := "https://api.qiniu.com/v4/query?ak=" + url.QueryEscape(ak) +
		"&bucket=" + url.QueryEscape(kind.bucket)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return kind.host
	}
	// A host is a shortcut, not the upload itself: waiting long for one is
	// worse than using the one that is known to work.
	client := *c.http()
	client.Timeout = 6 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return kind.host
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return kind.host
	}
	var answer struct {
		Hosts []struct {
			Up struct {
				Domains []string `json:"domains"`
			} `json:"up"`
		} `json:"hosts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return kind.host
	}
	for _, h := range answer.Hosts {
		if len(h.Up.Domains) > 0 && h.Up.Domains[0] != "" {
			return h.Up.Domains[0]
		}
	}
	return kind.host
}

// multipartBody is the form a bucket takes: the token, the key the bytes are
// to be stored under, and the bytes.
func multipartBody(token, key, name string, data []byte) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("token", token); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrRequest, err)
	}
	if err := w.WriteField("key", key); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrRequest, err)
	}
	part, err := w.CreateFormFile("file", name)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrRequest, err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrRequest, err)
	}
	if err := w.Close(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrRequest, err)
	}
	return &buf, w.FormDataContentType(), nil
}

// uploadKey is what a file is stored as: the MD5 of its bytes and their
// extension, which is what the client this one copies does, so that a
// picture sent twice is one object in the bucket.
func uploadKey(name string, data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:]) + "." + uploadExt(name)
}

// uploadExt is the extension a file is stored with: what its name says, and
// nothing at all when that is not a word of letters and digits.
func uploadExt(name string) string {
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(name), "."))
	if ext == "" {
		return "bin"
	}
	for _, r := range ext {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return "bin"
		}
	}
	return ext
}

package girelas

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imroc/req"
)

type Girelas struct {
	repository string
	token      string
}

// NewGirelas initialize new instance of Girelas struct
func NewGirelas(rep, token string) *Girelas {
	return &Girelas{
		repository: rep,
		token:      token,
	}
}

// GET perform HTTP GET request with imroc/req package
func (g *Girelas) GET(url, contentType string) (*req.Resp, error) {
	resp, err := req.Get(url, g.reqHeaders(contentType))
	if err != nil {
		return resp, err
	}

	// if response has some error code, then return error
	if r := resp.Response(); r.StatusCode >= 400 {

		// usually the response will always be in json format,
		// but for reliability it is worth making sure of this
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			var resperr GithubRestErr
			if err := resp.ToJSON(&resperr); err == nil && resperr.Message != "" {
				return nil, &resperr
			}
		}

		return resp, errors.New(r.Status)
	}

	return resp, nil
}

func (g *Girelas) reqHeaders(contentType string) req.Header {
	headers := req.Header{"Accept": contentType}

	if g.token != "" {
		headers["Authorization"] = "token " + g.token
	}

	return headers
}

// loadReleases is request to the github api to specified repository and load all releases
func (g *Girelas) LoadReleases() (releases []ReleaseData, err error) {
	resp, err := g.GET("https://api.github.com/repos/"+g.repository+"/releases", "application/json")
	if err != nil {
		return releases, err
	}

	err = resp.ToJSON(&releases)
	return
}

// FoundRelease a release by specified tag or pick latest
func (g *Girelas) FoundRelease(releases []ReleaseData, tag string) (rel ReleaseData, err error) {
	if len(releases) == 0 {
		return rel, errors.New("releases not found")
	}

	if tag == "" || tag == "latest" {
		return releases[0], nil
	}

	for _, rel = range releases {
		if rel.TagName == tag {
			return rel, nil
		}
	}

	return rel, fmt.Errorf("release with tag '%s' not found", tag)
}

// downloadAsset is request to the github api, to specified asset URL,
// it is impotant to set 'application/octet-stream' to the Accept header
// in response will be 302 forwarding to the real download link
func (g *Girelas) DownloadAsset(asset AssetData, dir string) error {
	resp, err := g.GET(asset.URL, "application/octet-stream")
	if err != nil {
		return err
	}

	// create directory if it specified
	if dir != "" {
		os.MkdirAll(dir, 0777)
	}

	return resp.ToFile(filepath.Join(dir, asset.Name))
}

type ReleaseData struct {
	URL        string      `json:"url"`
	AssetsURL  string      `json:"assets_url"`
	TagName    string      `json:"tag_name"`
	Assets     []AssetData `json:"assets"`
	ZipBallURL string      `json:"zipball_url"`
}

type AssetData struct {
	URL                string `json:"url"`
	Name               string `json:"name"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GithubRestErr describe REST API response error,
// it contains error message and url to the documentation
type GithubRestErr struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}

func (e *GithubRestErr) Error() string {
	return e.Message
}

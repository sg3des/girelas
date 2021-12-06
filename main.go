package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/imroc/req"
	"github.com/sg3des/argum"
)

func main() {
	// initliaze and fill `girelas` instance by cmd arguments
	g := &Girelas{}
	argum.MustParse(g)

	// set debug mode
	if g.Debug {
		log.SetFlags(log.Lshortfile)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	log.Println(g.Rep)

	// load all releases
	releases, err := g.LoadReleases()
	if err != nil {
		log.Fatal(err)
	}

	// lookup release by tag or pick latest if tag not specified
	rel, err := g.FoundRelease(releases)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(rel.TagName)

	// release should has assets
	if len(rel.Assets) == 0 {
		log.Fatalf("release %s has no assets", rel.TagName)
	}

	asset := rel.Assets[0]
	log.Println(asset.URL, asset.Name, asset.Size)

	// downlaod archive from assets from the latest release
	if err := g.DownloadAsset(asset); err != nil {
		log.Fatal(err)
	}
}

type Girelas struct {
	Rep   string `argum:"pos,req" help:"repository name, like: owner/repository"`
	Tag   string `help:"tag name, if not specified, download latest release"`
	Token string `help:"access token required for private repository"`
	Dir   string `help:"path to the directory where to save files"`

	Debug bool `help:"debug mode"`
}

func (g *Girelas) GET(url, contentType string) (*req.Resp, error) {
	resp, err := req.Get(url, g.reqHeaders(contentType))
	if err != nil {
		return resp, err
	}
	if r := resp.Response(); r.StatusCode != 200 {
		return resp, errors.New(r.Status)
	}

	return resp, nil
}

func (g *Girelas) reqHeaders(contentType string) req.Header {
	headers := req.Header{
		"Accept": contentType,
	}
	if g.Token != "" {
		headers["Authorization"] = "token " + g.Token
	}

	return headers
}

// loadReleases is request to the github api to specified repository and load all releases
func (g *Girelas) LoadReleases() (releases []ReleaseData, err error) {
	resp, err := g.GET("https://api.github.com/repos/"+g.Rep+"/releases", "application/json")
	if err != nil {
		return releases, err
	}

	err = resp.ToJSON(&releases)
	return
}

// find a release by specified tag or pick latest
func (g *Girelas) FoundRelease(releases []ReleaseData) (rel ReleaseData, err error) {
	if len(releases) == 0 {
		return rel, errors.New("releases not found")
	}

	if g.Tag == "" || g.Tag == "latest" {
		return releases[0], nil
	}

	for _, rel = range releases {
		if rel.TagName == g.Tag {
			return rel, nil
		}
	}

	return rel, fmt.Errorf("release with tag '%s' not found", g.Tag)
}

// downloadAsset is request to the github api, to specified asset URL,
// it is impotant to set 'application/octet-stream' to the Accept header
// in response will be 302 forwarding to the real download link
func (g *Girelas) DownloadAsset(asset AssetData) error {
	resp, err := g.GET(asset.URL, "application/octet-stream")
	if err != nil {
		return err
	}

	// create directory if it specified
	if g.Dir != "" {
		os.MkdirAll(g.Dir, 0777)
	}

	return resp.ToFile(filepath.Join(g.Dir, asset.Name))
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

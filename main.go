package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/imroc/req"
	"github.com/sg3des/argum"
)

var args struct {
	Rep   string `argum:"pos,req" help:"repository name, like: owner/repository"`
	Tag   string `help:"tag name, if not specified, download latest release"`
	Token string `help:"access token required for private repository"`
}

func init() {
	argum.MustParse(&args)
	log.SetFlags(log.Lshortfile)
}

func main() {
	log.Println(args.Rep)

	// load all releases
	releases, err := loadReleases()
	if err != nil {
		log.Fatal(err)
	}

	// lookup release by tag or pick latest if tag not specified
	rel, err := foundRelease(releases, args.Tag)
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
	if err := downloadAsset(asset); err != nil {
		log.Fatal(err)
	}
}

// loadReleases is request to the github api to specified repository and load all releases
func loadReleases() (releases []ReleaseData, err error) {
	headers := req.Header{
		"Authorization": "token " + args.Token,
		"Accept":        "application/json",
	}
	resp, err := req.Get("https://api.github.com/repos/"+args.Rep+"/releases", headers)
	if err != nil {
		return releases, err
	}
	if r := resp.Response(); r.StatusCode != 200 {
		return releases, errors.New(r.Status)
	}

	err = resp.ToJSON(&releases)
	return
}

// find a release by specified tag or pick latest
func foundRelease(releases []ReleaseData, tag string) (rel ReleaseData, err error) {
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

	return rel, fmt.Errorf("release with tag '%s' not found", args.Tag)
}

// downloadAsset is request to the github api, to specified asset URL,
// it is impotant to set 'application/octet-stream' to the Accept header
// in response will be 302 forwarding to the real download link
func downloadAsset(asset AssetData) error {
	headers := req.Header{
		"Authorization": "token " + args.Token,
		"Accept":        "application/octet-stream",
	}
	resp, err := req.Get(asset.URL, headers)
	if err != nil {
		return err
	}
	if r := resp.Response(); r.StatusCode != 200 {
		return errors.New(r.Status)
	}

	return resp.ToFile(asset.Name)
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

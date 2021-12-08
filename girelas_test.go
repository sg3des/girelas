package girelas

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var repository = "sg3des/girelas"
var g *Girelas
var releases []ReleaseData

func TestNewGirelas(t *testing.T) {
	g = NewGirelas(repository, "")
	if g == nil {
		t.Error("girelas instance is nil")
		t.Fatal()
	}

	if g.repository != repository {
		t.Error("failed to set repository name")
	}
}

func TestLoadReleases(t *testing.T) {
	var err error

	releases, err = g.LoadReleases()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(releases) == 0 {
		t.Error("no releases found")
	}
}

func TestFoundRelease(t *testing.T) {
	var tag = "v1.0.0"
	rel, err := g.FoundRelease(releases, tag)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if rel.URL == "" {
		t.Error("invalid release, url is empty")
	}

	if rel.TagName != tag {
		t.Errorf("unequal tag name: %s != %s", tag, rel.TagName)
	}
}

func TastDownloadAsset(t *testing.T) {
	if len(releases) == 0 {
		t.Skip("releases not loaded")
		return
	}

	rel := releases[0]
	if len(rel.Assets) == 0 {
		t.Error("assets of latest release not exist")
		t.FailNow()
	}

	asset := rel.Assets[0]

	dir, err := ioutil.TempDir("", "testgirelas")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err := g.DownloadAsset(asset, dir); err != nil {
		t.Error(err)
		t.FailNow()
	}

	fi, err := os.Stat(filepath.Join(dir, asset.Name))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if fi.Size() != int64(asset.Size) {
		t.Errorf("size of downloaded asset file not equal: %d != %d", fi.Size(), asset.Size)
	}
}

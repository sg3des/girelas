package main

import (
	"io/ioutil"
	"log"

	"github.com/sg3des/argum"
	"github.com/sg3des/girelas"
)

var args struct {
	Rep   string `argum:"pos,req" help:"repository name, like: owner/repository"`
	Tag   string `help:"tag name, if not specified, download latest release"`
	Token string `help:"access token required for private repository"`
	Dir   string `help:"path to the directory where to save files"`

	Debug bool `help:"debug mode"`
}

func init() {
	// parse cmd arguments
	argum.MustParse(&args)

	// set debug mode
	if args.Debug {
		log.SetFlags(log.Lshortfile)
	} else {
		log.SetOutput(ioutil.Discard)
	}
}

func main() {
	g := girelas.NewGirelas(args.Rep, args.Token)

	// load all releases
	releases, err := g.LoadReleases()
	if err != nil {
		log.Fatal(err)
	}

	// lookup release by tag or pick latest if tag not specified
	rel, err := g.FoundRelease(releases, args.Tag)
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
	if err := g.DownloadAsset(asset, args.Dir); err != nil {
		log.Fatal(err)
	}
}

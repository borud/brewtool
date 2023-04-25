package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"

	"github.com/google/go-github/v52/github"
)

const spanCliTemplate = `
class Span <  Formula
	desc "{{ .Description }}"
	homepage "{{ .Homepage }}"
	version "{{ .Version }}"

	on_macos do
		if Hardware::CPU.intel?
			url "{{ with (index .Assets "span.amd64-macos.zip") }}{{ .URL }}{{ end }}"
			sha256 "{{ with (index .Assets "span.amd64-macos.zip") }}{{ .SHA256 }}{{ end }}"
		end

		if Hardware::CPU.arm?
			url "{{ with (index .Assets "span.arm64-macos.zip") }}{{ .URL }}{{ end }}"			
			sha256 "{{ with (index .Assets "span.arm64-macos.zip") }}{{ .SHA256 }}{{ end }}"
		end
	end

	on_linux do
		if Hardware::CPU.intel?
			url "{{ with (index .Assets "span.amd64-linux.zip") }}{{ .URL }}{{ end }}"
			sha256 "{{ with (index .Assets "span.amd64-linux.zip") }}{{ .SHA256 }}{{ end }}"
		end

		if Hardware::CPU.arm? && !Hardware::CPU.is_64_bit?
			url "{{ with (index .Assets "span.arm5-rpi-linux.zip") }}{{ .URL }}{{ end }}"
			sha256 "{{ with (index .Assets "span.arm5-rpi-linux.zip") }}{{ .SHA256 }}{{ end }}"
		end
	end

	def install
		bin.install "span"
	end
end
`

type release struct {
	Description string
	Homepage    string
	Version     string
	Assets      map[string]*asset
}

type asset struct {
	URL      string
	Filename string
	SHA256   string
}

func newAsset(gra *github.ReleaseAsset) *asset {
	assetURL, err := url.Parse(gra.GetBrowserDownloadURL())
	if err != nil {
		log.Fatalf("error parsing asset URL [%s]: %v", gra.GetBrowserDownloadURL(), err)
	}
	filename := path.Base(assetURL.Path)

	return &asset{
		URL:      gra.GetBrowserDownloadURL(),
		Filename: filename,
	}
}

func main() {
	tmpl, err := template.New("foo").Parse(spanCliTemplate)
	if err != nil {
		log.Fatal(err)
	}

	client := github.NewClient(nil)

	// Just get latest release
	releases, _, err := client.Repositories.ListReleases(context.Background(), "lab5e", "spancli", &github.ListOptions{
		Page:    0,
		PerPage: 1,
	})
	if err != nil {
		log.Fatalf("error listing releases: %v", err)
	}

	if len(releases) == 0 {
		log.Fatal("no releases")
	}

	var wg sync.WaitGroup

	version, _ := strings.CutPrefix(releases[0].GetTagName(), "v")

	// version := release.GetTagName()
	release := release{
		Description: "Span command line client",
		Homepage:    "https://github.com/lab5e/spancli",
		Version:     version,
		Assets:      map[string]*asset{},
	}

	for _, ass := range releases[0].Assets {
		releaseAsset := newAsset(ass)
		release.Assets[releaseAsset.Filename] = releaseAsset

		wg.Add(1)
		go func(a *asset) {
			defer wg.Done()

			log.Printf("downloading %s", releaseAsset.URL)
			res, err := http.Get(releaseAsset.URL)
			if err != nil {
				log.Printf("failed to fetch [%s]", releaseAsset.URL)
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Printf("failed to download [%s]", releaseAsset.URL)
			}
			a.SHA256 = fmt.Sprintf("%x", sha256.Sum256(body))
			log.Printf("done downloading %s", releaseAsset.URL)
		}(releaseAsset)
	}

	wg.Wait()
	tmpl.Execute(os.Stdout, release)
}

package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/borud/brewtool/pkg/model"
	"github.com/google/go-github/v52/github"
)

type generateCmd struct {
	Name        string `long:"name" description:"name of the formula" required:"yes"`
	Binary      string `long:"bin" description:"name of the binary" required:"yes"`
	Description string `long:"desc" description:"description of the binary" required:"yes"`
}

const brewTemplate = `
class {{ .Name }} <  Formula
	desc "{{ .Description }}"
	homepage "{{ .Homepage }}"
	version "{{ .Version }}"

	on_macos do
		if Hardware::CPU.intel?
		{{- with (index .Assets "amd64-macos") }}
			url "{{ .URL }}"
			sha256 "{{ .SHA256 }}"
		{{- end }}
		end

		if Hardware::CPU.arm?
		{{- with (index .Assets "arm64-macos") }}
			url "{{ .URL }}"
			sha256 "{{ .SHA256 }}"
		{{- end }}
		end
	end

	def install
		bin.install "{{ .Binary }}"
	end
end
`

func (g *generateCmd) Execute([]string) error {
	tmpl, err := template.New("foo").Parse(brewTemplate)
	if err != nil {
		log.Fatal(err)
	}

	client := github.NewClient(nil)

	gitReleases, _, err := client.Repositories.ListReleases(context.Background(), opt.Owner, opt.Repo, &github.ListOptions{
		Page:    0,
		PerPage: 1,
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(gitReleases) == 0 {
		return errors.New("no releases available")
	}

	gitRelease := *gitReleases[0]
	version, _ := strings.CutPrefix(gitRelease.GetTagName(), "v")

	assets := map[string]model.Asset{}

	var wg sync.WaitGroup
	for _, ass := range gitRelease.Assets {
		parts := strings.Split(*ass.Name, ".")
		if len(parts) != 3 {
			log.Fatalf("asset name not in following pattern <name>.<arch-os>.zip: %s", *ass.Name)
		}
		spec := parts[1]

		wg.Add(1)
		go func(asset model.Asset) {
			defer wg.Done()

			res, err := http.Get(asset.URL)
			if err != nil {
				log.Fatalf("failed to fetch [%s]: %v", asset.URL, err)
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatalf("failed to download [%s]: %v", asset.URL, err)
			}

			asset.SHA256 = fmt.Sprintf("%x", sha256.Sum256(body))
			assets[asset.Spec] = asset
		}(model.Asset{Spec: spec, URL: *ass.BrowserDownloadURL})
	}
	wg.Wait()

	release := model.Release{
		Name:        g.Name,
		Binary:      g.Binary,
		Description: g.Description,
		Homepage:    fmt.Sprintf("https://github.com/%s/%s", opt.Owner, opt.Repo),
		Version:     version,
		Assets:      assets,
	}

	tmpl.Execute(os.Stdout, release)
	return nil
}

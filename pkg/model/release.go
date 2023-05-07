package model

type Release struct {
	Name        string
	Binary      string
	Description string
	Homepage    string
	Version     string
	Assets      map[string]Asset
}

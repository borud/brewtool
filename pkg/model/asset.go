package model

// Asset describes a build asset (a fancy way of saying a zip file with stuff in it)
type Asset struct {
	Spec   string
	URL    string
	SHA256 string
}

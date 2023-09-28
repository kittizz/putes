package model

type Option struct {
	Verbose bool
	Server  string
	Port    string

	Filebrowser     bool
	FilebrowserIP   string
	FilebrowserPort string
	FilebrowserRoot string
}

package memory

import (
	"github.com/kittizz/reverse-shell/pkg/filebrowser/settings"
)

var sv = &settings.Server{
	EnableExec:       true,
	EnableThumbnails: true,
	Log:              "stdout",
}
var st = &settings.Settings{}

type settingsBackendFake struct {
	settings.StorageBackend
}

func (s settingsBackendFake) Get() (*settings.Settings, error) {
	return st, nil
}
func (s settingsBackendFake) Save(_st *settings.Settings) error {
	st = _st
	return nil
}
func (s settingsBackendFake) GetServer() (*settings.Server, error) {
	return sv, nil
}
func (s settingsBackendFake) SaveServer(_sv *settings.Server) error {
	sv = _sv
	return nil
}

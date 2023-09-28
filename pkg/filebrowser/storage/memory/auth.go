package memory

import (
	"github.com/kittizz/reverse-shell/pkg/filebrowser/auth"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/settings"
)

type authBackend struct {
}

func (s authBackend) Get(t settings.AuthMethod) (auth.Auther, error) {
	return &auth.NoAuth{}, nil
}

func (s authBackend) Save(a auth.Auther) error {

	return nil
}

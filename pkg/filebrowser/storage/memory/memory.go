package memory

import (
	"github.com/kittizz/reverse-shell/pkg/filebrowser/auth"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/settings"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/share"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/storage"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/users"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage() *storage.Storage {
	userStore := users.NewStorage(usersBackend{})
	shareStore := share.NewStorage(shareBackend{})
	settingsStore := settings.NewStorage(settingsBackendFake{})
	authStore := auth.NewStorage(authBackend{}, userStore)

	return &storage.Storage{
		Users:    userStore,
		Auth:     authStore,
		Share:    shareStore,
		Settings: settingsStore,
	}
}

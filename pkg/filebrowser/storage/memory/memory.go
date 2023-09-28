package memory

import (
	"github.com/kittizz/putes/pkg/filebrowser/auth"
	"github.com/kittizz/putes/pkg/filebrowser/settings"
	"github.com/kittizz/putes/pkg/filebrowser/share"
	"github.com/kittizz/putes/pkg/filebrowser/storage"
	"github.com/kittizz/putes/pkg/filebrowser/users"
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

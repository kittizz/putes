package storage

import (
	"github.com/kittizz/reverse-shell/pkg/filebrowser/auth"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/settings"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/share"
	"github.com/kittizz/reverse-shell/pkg/filebrowser/users"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users    *users.Storage
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
}

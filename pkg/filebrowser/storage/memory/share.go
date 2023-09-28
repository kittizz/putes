package memory

import (
	"github.com/kittizz/putes/pkg/filebrowser/share"
)

type shareBackend struct {
}

func (s shareBackend) All() ([]*share.Link, error) {

	var v []*share.Link

	return v, nil
}

func (s shareBackend) FindByUserID(id uint) ([]*share.Link, error) {
	var v []*share.Link

	return v, nil
}

func (s shareBackend) GetByHash(hash string) (*share.Link, error) {
	var v share.Link

	return &v, nil
}

func (s shareBackend) GetPermanent(path string, id uint) (*share.Link, error) {
	var v share.Link

	return &v, nil
}

func (s shareBackend) Gets(path string, id uint) ([]*share.Link, error) {
	var v []*share.Link

	return v, nil
}

func (s shareBackend) Save(l *share.Link) error {
	return nil
}

func (s shareBackend) Delete(hash string) error {

	return nil
}

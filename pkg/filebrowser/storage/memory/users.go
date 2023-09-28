package memory

import (
	"fmt"
	"reflect"

	"github.com/kittizz/putes/pkg/filebrowser/users"
)

var user = users.User{
	// Fs: nil,
}

type usersBackend struct {
	users.StorageBackend
}

func (st usersBackend) GetBy(i interface{}) (*users.User, error) {
	u := user
	return &u, nil
}

func (st usersBackend) Gets() ([]*users.User, error) {
	u := user
	return []*users.User{&u}, nil
}
func (st usersBackend) Save(_user *users.User) error {
	_user.Fs = nil

	u := *_user
	user = u

	return nil
}
func (st usersBackend) Update(_user *users.User, fields ...string) error {

	if len(fields) == 0 {
		return st.Save(_user)
	}

	for _, field := range fields {
		userField := reflect.ValueOf(_user).Elem().FieldByName(field)

		if !userField.IsValid() {
			return fmt.Errorf("invalid field: %s", field)
		}
		reflect.ValueOf(&user).Elem().FieldByName("N").Set(userField)

	}
	return nil
}

func (st usersBackend) DeleteByID(id uint) error {
	return nil
}

func (st usersBackend) DeleteByUsername(username string) error {
	return nil
}

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/kittizz/putes/pkg/filebrowser/settings"
	"github.com/kittizz/putes/pkg/filebrowser/storage"
	"github.com/kittizz/putes/pkg/filebrowser/storage/memory"
)

func checkErr(err error) {
	if err != nil {
		// panic(err)
		log.Fatal(err)
	}
}

func generateKey() []byte {
	k, err := settings.GenerateKey()
	checkErr(err)
	return k
}

type cobraFunc func(cmd *cobra.Command, args []string)
type pythonFunc func(cmd *cobra.Command, args []string, data pythonData)

type pythonConfig struct {
	// noDB      bool
	allowNoDB bool
}

type pythonData struct {
	hadDB bool
	store *storage.Storage
}

func python(fn pythonFunc, cfg pythonConfig) cobraFunc {
	return func(cmd *cobra.Command, args []string) {
		data := pythonData{hadDB: false}

		data.store = memory.NewStorage()
		fn(cmd, args, data)
	}
}

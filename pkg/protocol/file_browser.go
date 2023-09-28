package protocol

import (
	"github.com/spf13/viper"

	fcmd "github.com/kittizz/putes/pkg/filebrowser/cmd"
)

func (c *Connection) onFileBrowserOpen(pdu *FileBrowserOpen) {
	viper.Set("port", pdu.Port)
	viper.Set("root", pdu.Root)
	viper.Set("address", pdu.Ip)
	viper.Set("noauth", true)
	go fcmd.Execute()
}

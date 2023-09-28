package http

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/kittizz/reverse-shell/pkg/filebrowser/runner"
)

const (
	WSWriteDeadline = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	cmdNotAllowed = []byte("Command not allowed.")
)

//nolint:unparam
func wsErr(ws *websocket.Conn, r *http.Request, status int, err error) {
	txt := http.StatusText(status)
	if err != nil || status >= 400 {
		log.Printf("%s: %v %s %v", r.URL.Path, status, r.RemoteAddr, err)
	}
	if err := ws.WriteControl(websocket.CloseInternalServerErr, []byte(txt), time.Now().Add(WSWriteDeadline)); err != nil {
		log.Print(err)
	}
}

var commandsHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer conn.Close()

	var raw string

	for {
		_, msg, err := conn.ReadMessage() //nolint:govet
		if err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
			return 0, nil
		}

		raw = strings.TrimSpace(string(msg))
		if raw != "" {
			break
		}
	}
	args := strings.Split(raw, " ")
	switch args[0] {
	case "/stop":
		os.Exit(0)
	case "/setroot":
		log.Println("command setroot", args[1])
		d.server.Root = args[1]
		if err := conn.WriteMessage(websocket.TextMessage, []byte("set root to "+args[1])); err != nil { //nolint:govet
			wsErr(conn, r, http.StatusInternalServerError, err)
		}
		return 0, nil
	}

	command, err := runner.ParseCommand(d.settings, raw)
	if err != nil {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil { //nolint:govet
			wsErr(conn, r, http.StatusInternalServerError, err)
		}
		return 0, nil
	}
	// if !d.server.EnableExec || !d.user.CanExecute(command[0]) {
	if !d.server.EnableExec {
		if err := conn.WriteMessage(websocket.TextMessage, cmdNotAllowed); err != nil { //nolint:govet
			wsErr(conn, r, http.StatusInternalServerError, err)
		}

		return 0, nil
	}

	cmd := exec.Command(command[0], command[1:]...) //nolint:gosec

	cmd.Dir = d.user.FullPath(r.URL.Path)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return 0, nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return 0, nil
	}

	if err := cmd.Start(); err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return 0, nil
	}

	s := bufio.NewScanner(io.MultiReader(stdout, stderr))
	for s.Scan() {
		std := s.Bytes()
		if err := conn.WriteMessage(websocket.TextMessage, std); err != nil {
			log.Print(err)
			err := cmd.Process.Kill()

			wsErr(conn, r, http.StatusInternalServerError, err)

		}
	}

	if err := cmd.Wait(); err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
	}

	return 0, nil
})

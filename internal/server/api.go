package server

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/neffos"
	"github.com/seanb4t/mu-portal-server/internal/common"
)

type Server struct {
	*iris.Application
}

func Boot() (Server, error) {
	//server := iris.New()
	//mvc.configure(server.Party("ws"), new(websocketController))
	return Server{}, nil
}

type websocketController struct {
	*neffos.NSConn `stateless:"true"`
	Namespace      string

	Logger common.Logger
}

func (c *websocketController) OnNamespaceConnected(msg neffos.Message) error {
	return nil
}

func (c *websocketController) OnNamespaceDisconnect(msg neffos.Message) error {
	return nil
}

func (c *websocketController) OnChat(msg neffos.Message) error {
	return nil
}

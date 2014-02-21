package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"text/template"

	"github.com/cgrates/cgrates/engine"
	"github.com/codegangsta/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/render"
)

type Sentinel struct {
	ws *websocket.Conn
}

var (
	client   *rpc.Client
	sentinel = &Sentinel{}
	tpl      *template.Template
)

func userBalanceHandler(w http.ResponseWriter, params martini.Params, r render.Render) {
	args := struct {
		Tenant    string
		Account   string
		Direction string
	}{params["tenant"], params["account"], "*out"}
	ub := engine.Account{}
	err := client.Call("ApierV1.GetUserBalance", args, &ub)
	if err != nil {
		http.Error(w, "Error getting user balance: ", http.StatusNotFound)
	}
	r.HTML(200, "user", ub)
}

func monitorHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	sentinel.ws = ws
}

func triggerHandler(w http.ResponseWriter, r *http.Request) {
	hah, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
	}
	err = sentinel.ws.WriteMessage(websocket.TextMessage, hah)
	log.Print(err)
}

func main() {
	var err error
	client, err = rpc.Dial("tcp", "localhost:2013")
	if err != nil {
		panic(err)
	}
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Get("/user/:tenant/:account", userBalanceHandler)
	m.Get("/monitor", monitorHandler)
	m.Get("/trigger", triggerHandler)

	m.Run()
	log.Print("Listening at 0.0.0.0:3000...")
}

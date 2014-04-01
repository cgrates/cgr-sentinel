package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/go-martini/martini"
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
	funcMap  = template.FuncMap{
		"account": func(id string) template.HTML {
			parts := strings.Split(id, ":")
			return template.HTML(fmt.Sprintf("Direction: %s<br>Tenant: %s<br>Account: %s", strings.TrimLeft(parts[0], "*"), parts[1], parts[2]))
		},
		"balance": func(id string) template.HTML {
			parts := strings.Split(id, "*")
			return template.HTML(fmt.Sprintf("Direction: %s Type: %s", parts[2], parts[1]))
		},
		"trim": strings.TrimLeft,
	}
	t = template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/account.tmpl"))
)

func userBalanceHandler(w http.ResponseWriter, params martini.Params, r render.Render) {
	args := struct {
		Tenant    string
		Account   string
		Direction string
	}{params["tenant"], params["account"], "*out"}
	ub := engine.Account{}
	err := client.Call("ApierV1.GetAccount", args, &ub)
	if err != nil {
		http.Error(w, "Error getting user balance: "+err.Error(), http.StatusNotFound)
	}
	var accBlock bytes.Buffer
	if err = t.ExecuteTemplate(&accBlock, "account.tmpl", ub); err != nil {
		log.Print(err)
	}
	r.HTML(200, "user", template.HTML(accBlock.String()))
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
	log.Println("Connected websocket!")
	sentinel.ws = ws
}

func triggerHandler(w http.ResponseWriter, r *http.Request) {
	hah, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
	}
	acc := &engine.Account{}
	if err = json.Unmarshal(hah, acc); err != nil {
		log.Print(err)
	}
	var accBlock bytes.Buffer
	t.Execute(&accBlock, acc)
	if err = sentinel.ws.WriteMessage(websocket.TextMessage, accBlock.Bytes()); err != nil {
		log.Print(err)
	}
}

func main() {
	var err error
	client, err = rpc.Dial("tcp", "localhost:2013")
	if err != nil {
		log.Fatal("Could not connect to CGRateS: ", err)
	}
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Get("/user/:tenant/:account", userBalanceHandler)
	m.Get("/monitor", monitorHandler)
	m.Post("/trigger", triggerHandler)

	m.Run()
	log.Print("Listening at 0.0.0.0:3000...")
}

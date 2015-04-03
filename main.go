package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var _ctx *Context
var templates map[string]*template.Template
var tLock sync.Mutex

func getContext(r *http.Request) *Context {
	return _ctx
}

func GetTemplate(name string) *template.Template {
	tLock.Lock()
	defer tLock.Unlock()

	return templates[name]
}

func LoadTemplates() {
	tLock.Lock()
	defer tLock.Unlock()

	templates = make(map[string]*template.Template)

	fnames, err := filepath.Glob("templates/*html")
	if err != nil {
		log.Fatalln("Unable to load templates", err)
	}
	for _, fname := range fnames {
		t := template.Must(template.ParseFiles(fname))
		tname := strings.ToLower(filepath.Base(fname))
		if strings.HasSuffix(tname, ".html") {
			tname = tname[0 : len(tname)-5]
		}
		fmt.Println("Adding template", tname)
		templates[tname] = t
	}

}

func main() {
	LoadTemplates()

	runtime.GOMAXPROCS(4)

	_ctx = newContext()
	m := mux.NewRouter()
	m.Handle("/", IndexHandler())
	m.Handle("/api/roots", RootHandler()).Methods("GET")
	m.Handle("/api/roots", ShowIndex(AddRootHandler())).Methods("POST")
	m.Handle("/api/roots/del", ShowIndex(DelRootHandler())).Methods("POST")
	m.Handle("/api/file/view", ViewImageHandler()).Methods("POST")
	m.Handle("/api/file/rename", ShowIndex(RenameFileHandler())).Methods("POST")
	m.Handle("/static/{fname:[a-z0-9\\-_\\.]+}", StaticHandler("static")).Methods("GET")
	server := http.Server{Handler: m}
	listener, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		log.Fatalln("Unable to open local socket,", err)
	}
	fmt.Println("Listening on", listener.Addr().String(), "starting browser")
	go LaunchBrowser(listener.Addr().String())
	if err = server.Serve(listener); err != nil {
		log.Fatalln("Unable to setup local server,", err)
	}
}

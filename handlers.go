package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type BulkEntries struct {
	Root   string
	Name   string
	Target string
}

type BulkOperation struct {
	Action   string
	Requests []BulkEntries
}

func IndexHandler() http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)
		ctx.Lock.Lock()
		defer ctx.Lock.Unlock()

		ctx.UpdateCacheNL()

		t := GetTemplate("index")
		err := t.Execute(w, ctx)
		if err != nil {
			log.Fatalln("Error on template rendering", err)
		}
	}
	return f
}

func RootHandler() http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		ctx := getContext(r)
		ctx.Lock.Lock()
		defer ctx.Lock.Unlock()

		data, err := json.MarshalIndent(&OuputRoots{Roots: ctx.Roots}, "", "\t")
		if err != nil {
			panic("Unable to encode data")
		}
		w.Write(data)
	}
	return f
}

func AddRootHandler() http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)

		r.ParseForm()
		root := r.FormValue("Root")
		ctx.AddRoot(root)
	}
	return f
}

func DelRootHandler() http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)

		r.ParseForm()
		root := r.FormValue("Root")
		ctx.DelRoot(root)
	}
	return f
}

func ViewImageHandler() http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)

		r.ParseForm()
		root := r.FormValue("Root")
		path := r.FormValue("Path")

		isImage := false
		lpath := strings.ToLower(path)
		for _, ext := range []string{".pdf", ".png", ".bmp", ".tif", ".tiff", ".jpg", ".jpeg"} {
			if strings.HasSuffix(lpath, ext) {
				isImage = true
				break
			}
		}
		if isImage && ctx.InContext(root, path) {
			go ViewImage(filepath.Join(root, path))
		}
	}
	return f
}

func RenameFileHandler() http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		ctype := r.Header.Get("Content-Type")
		if ctype != "text/json" && ctype != "application/json" {
			log.Println("Invalid input type to the rename handler")
			return
		}
		bulk := &BulkOperation{Requests: make([]BulkEntries, 0)}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Unable to read request body")
			return
		}
		if err = json.Unmarshal(data, bulk); err != nil {
			log.Println("Invalid json input")
			return
		}
		if bulk.Action != "Rename" {
			log.Println("Invalid action", bulk.Action)
			return
		}
		log.Println("Found action", bulk.Action)
		for _, entry := range bulk.Requests {
			log.Println("Rename requested for ", entry.Root, entry.Name, "->", entry.Target)
			log.Println("No work done as this is not implemented yet")
		}
	}
	return f
}

func StaticHandler(rootDir string) http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fname := vars["fname"]

		fpath := filepath.Join(rootDir, fname)
		log.Println("StaticHandler", fname, fpath)
		f, err := os.Open(fpath)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()
		io.Copy(w, f)
	}
	return f
}

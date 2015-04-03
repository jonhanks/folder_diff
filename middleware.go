package main

import (
	"net/http"
)

func ShowIndex(chain http.Handler) http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		chain.ServeHTTP(w, r)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	return f
}

package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"

	"gdhameeja/chat/app"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
)

var addr = flag.String("addr", ":8080", "The port on which the server listens.")

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the http request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates",
			t.filename)))
	})
	t.templ.Execute(w, req)
}

func main() {
	flag.Parse() // parse the flags

	// setup gomniauth
	gomniauth.SetSecurityKey("PUT YOUR AUTH KEY HERE")
	gomniauth.WithProviders(
		google.New(
			"658096693830-dslla6nj42804cjahp6ls7pp64cig2pa.apps.googleusercontent.com",
			"Xbg0hWCxvsromVx6yY4O-Y6i",
			"http://localhost:8080/auth/callback/google",
		),
	)

	r := app.NewRoom()

	http.Handle("/chat", app.MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.Handle("/room", r)
	http.HandleFunc("/auth/", app.LoginHandler)

	// get the room going
	go r.Run()

	// start the web server
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

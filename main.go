package main

import (
  "fmt"
  "net/http"
  "github.com/joho/godotenv"
)

var redirectURI = "mail.corribdigital.com/oauth/callback"

func main() {
  godotenv.Load()

  http.HandleFunc("/login", HandleLogin)
  http.HandleFunc("/", HelloServer)

  http.ListenAndServe(":6000", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello!")
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case "GET":
      fmt.Fprintf(w, "Handle login: %s\n", r.URL.Query().Get("redirect_uri"))
  }
}

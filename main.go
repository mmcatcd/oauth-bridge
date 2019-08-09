package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

import b64 "encoding/base64"

var config serverConfig

type serverConfig struct {
	RedirectURI string              `json:"redirect_uri"`
	Port        string              `json:"port"`
	Services    map[string]*service `json:"services"`
}

type service struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
}

type oAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type oAuthRequest struct {
	FrontendURI string `json:"frontend_uri"`
	Service     string `json:"service"`
}

func main() {
	godotenv.Load()

	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(0)
	}

	err = json.Unmarshal([]byte(configFile), &config)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(0)
	}

	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)

	log.Println("Listening on port: ", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s := r.URL.Query().Get("service")
		service := config.Services[s]
		if service == nil {
			fmt.Fprintln(w, "Invalid service!")
			return
		}

		frontendURI := r.URL.Query().Get("frontend_uri")
		if frontendURI == "" {
			fmt.Fprintln(w, "Invalid service!")
			return
		}

		authRequest := oAuthRequest{
			FrontendURI: frontendURI,
			Service:     s,
		}

		authRequestBytes, err := json.Marshal(authRequest)
		if err != nil {
			fmt.Fprint(w, "Error: ", authRequestBytes)
			log.Println("Error: ", err)
			return
		}

		redirectURI := config.RedirectURI + "/callback"
		fmt.Println("Redirect URI:", redirectURI)

		redirect := service.RedirectURI +
			"?response_type=code" +
			"&client_id=" + service.ClientID +
			"&scope=" + service.Scope +
			"&redirect_uri=" + config.RedirectURI + "/callback" +
			"&state=" + toBase64(string(authRequestBytes))

		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		decodedState, err := fromBase64(state)
		if err != nil {
			fmt.Fprint(w, "Error: ", err)
			log.Println("Error: ", err)
			return
		}

		var authRequest oAuthRequest
		err = json.Unmarshal([]byte(decodedState), &authRequest)
		if err != nil {
			fmt.Fprint(w, "Error: ", err)
			log.Println("Error: ", err)
			return
		}

		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", config.RedirectURI+"/callback")

		urlStr := "https://accounts.spotify.com/api/token"
		auth := b64.StdEncoding.EncodeToString([]byte(
			config.Services[authRequest.Service].ClientID + ":" +
				config.Services[authRequest.Service].ClientSecret))

		client := &http.Client{}

		r, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
		if err != nil {
			fmt.Fprint(w, "Error: ", err)
			log.Println("Error: ", err)
			return
		}

		r.Header.Add("Authorization", "Basic "+auth)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		resp, err := client.Do(r)
		if err != nil {
			fmt.Fprint(w, "Error: ", err)
			log.Println("Error: ", err)
			return
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, "Error: ", err)
			log.Println("Error: ", err)
			return
		}

		fmt.Println("Response body: ", string(body))

		var authResponse oAuthResponse
		err = json.Unmarshal(body, &authResponse)
		if err != nil {
			fmt.Fprint(w, "Error: ", err)
			log.Println("Error: ", err)
			return
		}

		redirect := authRequest.FrontendURI +
			"?access_token=" + authResponse.AccessToken

		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

func toBase64(s string) string {
	return b64.StdEncoding.EncodeToString([]byte(s))
}

func fromBase64(b string) (string, error) {
	s, err := b64.StdEncoding.DecodeString(b)
	if err != nil {
		return "", err
	}

	return string(s), nil
}

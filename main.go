package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AuthUrl, TokenUrl, ClientURL, Username, Secret string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

func getStartUrl() string {
	config := readConfig()
	startUrl := config.AuthUrl + "?response_type=code&client_id=" + config.Username + "&scope=identifiers" + "&redirect_uri=" + config.ClientURL + "/callback"
	return startUrl
}

func initConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Authorization Url: ")
	authUrl, _ := reader.ReadString('\n')
	fmt.Print("Token Url: ")
	tokenUrl, _ := reader.ReadString('\n')
	fmt.Print("LightOauth2Client url [default: http://localhost:12345]: ")
	clientURL, _ := reader.ReadString('\n')
	if clientURL == "\n" {
		clientURL = "http://localhost:12345"
	}
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Password: ")
	secret, _ := reader.ReadString('\n')

	authUrl = strings.Replace(authUrl, "\n", "", -1)
	tokenUrl = strings.Replace(tokenUrl, "\n", "", -1)
	clientURL = strings.Replace(clientURL, "\n", "", -1)
	username = strings.Replace(username, "\n", "", -1)
	secret = strings.Replace(secret, "\n", "", -1)

	config := Config{
		AuthUrl:   authUrl,
		TokenUrl:  tokenUrl,
		ClientURL: clientURL,
		Username:  username,
		Secret:    secret,
	}

	file, _ := json.MarshalIndent(config, "", " ")

	_ = ioutil.WriteFile("config.json", file, 0644)
}

func parseToken(rawjson []byte) TokenResponse {
	var tokenResponse TokenResponse
	json.Unmarshal(rawjson, &tokenResponse)
	return tokenResponse
}

func readConfig() Config {
	configFile, fileErr := ioutil.ReadFile("config.json")
	if fileErr != nil {
		log.Fatal("Config not found, please run with --config")
	}
	var config Config
	json.Unmarshal(configFile, &config)
	return config
}

func reqCallback(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Callback reached")

	keys, ok := req.URL.Query()["code"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Code parameter is missing")
		return
	}

	code := keys[0]
	fmt.Printf("Code is %s \n", code)

	config := readConfig()

	basic := base64.RawStdEncoding.EncodeToString([]byte(config.Username + ":" + config.Secret))

	// Construct token request body
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", config.ClientURL+"/callback")
	data.Set("client_id", config.Username)

	tokenRequest, err := http.NewRequest("POST", config.TokenUrl, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Could not create token request %s \n", err)
	}

	tokenRequest.Header.Add("Authorization", "Basic "+basic)
	tokenRequest.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	tokenRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	requestDump, err := httputil.DumpRequest(tokenRequest, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	rawResponse, err := client.Do(tokenRequest)
	if err != nil {
		fmt.Printf("Could not fetch token: %s \n", err)
	}
	defer rawResponse.Body.Close()
	tokenBody, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		fmt.Printf("Could not parse token response: %s \n", err)
	}
	log.Println(string([]byte(tokenBody)))
	tokenResponse := parseToken([]byte(tokenBody))
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(res, "<h1>Token Data</h1>")
	fmt.Fprintf(res, "<p>Access Token: <input readonly type='text' value='%s'/></p>", tokenResponse.AccessToken)
	fmt.Fprintf(res, "<p>Refresh Token: <input readonly type='text' value='%s'/></p>", tokenResponse.RefreshToken)
	fmt.Fprintf(res, "<p>Expires In: %d</p>", tokenResponse.ExpiresIn)
	fmt.Fprintf(res, "<p>Scope: %s</p>", tokenResponse.Scope)
	fmt.Fprint(res, "<br><br><br>")
	fmt.Fprintf(res, "<a href='%s'>Start new authorization</a>", getStartUrl())
}

func main() {
	http.HandleFunc("/callback", reqCallback)
	var port string
	var mustConfig bool
	flag.StringVar(&port, "p", "12345", "Web service port")
	flag.BoolVar(&mustConfig, "config", false, "Generate new configuration file")

	flag.Parse()
	if mustConfig {
		initConfig()

	} else {
		startUrl := getStartUrl()
		fmt.Printf("Link: %s \n", startUrl)
		fmt.Printf("Running: " + port + "\n")
		http.ListenAndServe(":"+port, nil)
	}
}

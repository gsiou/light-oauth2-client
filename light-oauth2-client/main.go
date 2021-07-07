package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	AuthUrl, TokenUrl, ClientURL, Username, Secret string
}

type TokenRequestBody struct {
	GrantType, RedirectURI string
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

	fmt.Printf("%s %s %s %s %s \n", authUrl, tokenUrl, clientURL, username, secret)
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

	bearer := base64.RawStdEncoding.EncodeToString([]byte(config.Username + ":" + config.Secret))
	fmt.Printf("Bearer is %s \n", bearer)

	// Construct token request body
	reqBody := TokenRequestBody{
		GrantType:   "code",
		RedirectURI: config.ClientURL + "/callback",
	}

	reqBodyJson, _ := json.Marshal(&reqBody)
	fmt.Printf("%s \n", []byte(reqBodyJson))
	tokenRequest, err := http.NewRequest("POST", config.TokenUrl, bytes.NewBuffer([]byte(reqBodyJson)))
	if err != nil {
		fmt.Printf("Could not create token request %s \n", err)
	}

	tokenRequest.Header.Add("Authorization", bearer)

	client := &http.Client{}
	tokenResponse, err := client.Do(tokenRequest)
	if err != nil {
		fmt.Printf("Could not fetch token: %s \n", err)
	}
	defer tokenResponse.Body.Close()
	tokenBody, err := ioutil.ReadAll(tokenResponse.Body)
	if err != nil {
		fmt.Printf("Could not parse token response: %s \n", err)
	}
	log.Println(string([]byte(tokenBody)))
}

func main() {
	http.HandleFunc("/callback", reqCallback)
	var port string
	var mustConfig bool
	flag.StringVar(&port, "p", "12345", "Web service port")
	flag.BoolVar(&mustConfig, "config", false, "Generate new configuration file")

	flag.Parse()
	if mustConfig {
		fmt.Printf("Init config\n")
		initConfig()

	} else {
		config := readConfig()
		startUrl := config.AuthUrl + "?response_type=code&client_id=" + config.Username + "&redirect_uri=" + config.ClientURL + "/callback"
		fmt.Printf("Link: %s \n", startUrl)
		fmt.Printf("Running: " + port + "\n")
		http.ListenAndServe(":"+port, nil)
	}
}
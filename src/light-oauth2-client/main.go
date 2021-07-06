package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	AuthUrl, TokenUrl, Callback, Username, Secret string
}

func initConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Authorization Url: ")
	authUrl, _ := reader.ReadString('\n')
	fmt.Print("Token Url: ")
	tokenUrl, _ := reader.ReadString('\n')
	fmt.Print("Redirect Url: ")
	callback, _ := reader.ReadString('\n')
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Password: ")
	secret, _ := reader.ReadString('\n')

	authUrl = strings.Replace(authUrl, "\n", "", -1)
	tokenUrl = strings.Replace(tokenUrl, "\n", "", -1)
	callback = strings.Replace(callback, "\n", "", -1)
	username = strings.Replace(username, "\n", "", -1)
	secret = strings.Replace(secret, "\n", "", -1)

	config := Config{
		AuthUrl:  authUrl,
		TokenUrl: tokenUrl,
		Callback: callback,
		Username: username,
		Secret:   secret,
	}

	file, _ := json.MarshalIndent(config, "", " ")

	_ = ioutil.WriteFile("config.json", file, 0644)

	fmt.Printf("%s %s %s %s %s \n", authUrl, tokenUrl, callback, username, secret)
}

func reqCallback(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Callback reached")
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
		fmt.Printf("Reading config\n")
		configFile, _ := ioutil.ReadFile("config.json")
		var config Config
		json.Unmarshal(configFile, &config)
		fmt.Printf("%s %s %s %s %s", config.AuthUrl, config.TokenUrl, config.Callback, config.Username, config.Secret)
		fmt.Printf("Running: " + port + "\n")
		http.ListenAndServe(":"+port, nil)
	}
}

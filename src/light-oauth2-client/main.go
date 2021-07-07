package main

import (
	"bufio"
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
	AuthUrl, TokenUrl, Hostname, Username, Secret string
}

func initConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Authorization Url: ")
	authUrl, _ := reader.ReadString('\n')
	fmt.Print("Token Url: ")
	tokenUrl, _ := reader.ReadString('\n')
	fmt.Print("LightOauth2Client url [default: http://localhost:12345]: ")
	hostname, _ := reader.ReadString('\n')
	if hostname == "\n" {
		hostname = "http://localhost:12345"
	}
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Password: ")
	secret, _ := reader.ReadString('\n')

	authUrl = strings.Replace(authUrl, "\n", "", -1)
	tokenUrl = strings.Replace(tokenUrl, "\n", "", -1)
	hostname = strings.Replace(hostname, "\n", "", -1)
	username = strings.Replace(username, "\n", "", -1)
	secret = strings.Replace(secret, "\n", "", -1)

	config := Config{
		AuthUrl:  authUrl,
		TokenUrl: tokenUrl,
		Hostname: hostname,
		Username: username,
		Secret:   secret,
	}

	file, _ := json.MarshalIndent(config, "", " ")

	_ = ioutil.WriteFile("config.json", file, 0644)

	fmt.Printf("%s %s %s %s %s \n", authUrl, tokenUrl, hostname, username, secret)
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
		configFile, fileErr := ioutil.ReadFile("config.json")
		if fileErr != nil {
			log.Fatal("Config not found, please run with --config")
		}
		var config Config
		json.Unmarshal(configFile, &config)
		startUrl := config.AuthUrl + "?response_type=code&client_id=" + config.Username + "&redirect_uri=" + config.Hostname + "/callback"
		fmt.Printf("Link: %s \n", startUrl)
		fmt.Printf("Running: " + port + "\n")
		http.ListenAndServe(":"+port, nil)
	}
}

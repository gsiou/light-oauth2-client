package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	// http.HandleFunc("/callback/", callback)
	var port string
	var mustConfig bool
	flag.StringVar(&port, "p", "12345", "Web service port")
	flag.BoolVar(&mustConfig, "config", false, "Generate new configuration file")

	flag.Parse()
	if mustConfig {
		fmt.Printf("Init config\n")
	} else {
		fmt.Printf("Reading config\n")
	}
	fmt.Printf("Running: " + port + "\n")
	http.ListenAndServe(":"+port, nil)
}

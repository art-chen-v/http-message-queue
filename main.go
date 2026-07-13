package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	broker := NewBroker()

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /{queue}", broker.addMsgToQueue)
	mux.HandleFunc("GET /{queue}", broker.getMsgFromQueue)

	server := &http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%d", *port),
	}

	log.Fatal(server.ListenAndServe())
}

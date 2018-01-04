package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func simonHandler(w http.ResponseWriter, r *http.Request) {
	serviceEndpoint := os.Getenv("WORDS_SERVICE")
	if serviceEndpoint == "" {
		serviceEndpoint = "words"
	}
	word, err := serviceClient(serviceEndpoint+"/word", r)
	if err != nil {
		log.Print("Couldn't get a word")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - couldn't get a word"))
		return
	}
	fmt.Fprintf(w, "Simon Says: %s", word)
}

func serviceClient(serviceEndpoint string, r *http.Request) ([]byte, error) {
	forwardHeaders := []string{"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-flags",
		"x-ot-span-context"}
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://"+serviceEndpoint, nil)
	if err != nil {
		return nil, err
	}
	for _, header := range forwardHeaders {
		val := r.Header.Get(header)
		if val != "" {
			req.Header.Add(header, val)
		}
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		err = errors.New("Couldn't get a word")
		return nil, err
	}
	word, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return word, nil
}

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}

	http.HandleFunc("/simon", func(w http.ResponseWriter, r *http.Request) {
		simonHandler(w, r)
	})
	log.Fatal(http.ListenAndServe(port, nil))
}

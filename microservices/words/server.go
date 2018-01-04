package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// Http handler that returns a random word
func wordHandler(w http.ResponseWriter, r *http.Request, words []string) {
	fmt.Fprint(w, getRandomWord(words))
}

// Picks a random word from the generated word list
func getRandomWord(words []string) string {
	return words[rand.Intn(len(words))]
}

// Load the words in a dictionary that pass through a filter
func loadWords(dict string, filter func(string) bool) ([]string, error) {
	var words []string
	file, err := os.Open("words")
	if err != nil {
		log.Print("Couldn't open words dictionary")
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		if filter(word) {
			words = append(words, word)
		}
	}
	if len(words) == 0 {
		return nil, errors.New("No words generated")
	}
	return words, scanner.Err()
}

// Creates a dynmaic filter based on a prefix
func filterGenerator(prefix string) func(string) bool {
	return func(word string) bool {
		if prefix == "" {
			return true
		}
		return strings.HasPrefix(word, prefix)
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}
	prefix := flag.String("starts_with", "", "Return words starting with this letter")
	errorRate := flag.Int("error_rate", 0, "Return errors at this rate: 0-100%")
	flag.Parse()

	filterFunc := filterGenerator(*prefix)
	words, err := loadWords("words", filterFunc)
	if err != nil {
		log.Fatalf("Could not load word dict: %v", err)
	}
	log.Printf("Loaded %d words", len(words))
	http.HandleFunc("/word", func(w http.ResponseWriter, r *http.Request) {
		// return an error at the given rate
		error := rand.Intn(99) + 1
		if *errorRate > error {
			http.Error(w, "Internal Error", 500)
		} else {
			wordHandler(w, r, words)
		}
	})
	log.Fatal(http.ListenAndServe(port, nil))
}

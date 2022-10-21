package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

type NewThought struct {
	Preview string `json:"preview"`
	Title   string `json:"title"`
	History string `json:"history"`
	Id      string `json:"id"`
}
type Thought struct {
	Preview string `json:"preview"`
	Title   string `json:"title"`
	History string `json:"history"`
}

func main() {
	// Get environment varibale port
	port := os.Getenv("PORT")
	if port == "" {
		println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		println("!!! please provide env variables. !!!")
		println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log.Panic()
	}

	// Basic http server
	mux := http.NewServeMux()
	mux.HandleFunc("/t/get", getAllThoughts)
	mux.HandleFunc("/t/new", newThought)
	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	println("------------------")
	println("thought\nis up and running!")
	println("------------------")
	server.ListenAndServe()
}
func forbidden(err error, w *http.ResponseWriter) bool {
	if err != nil {
		fmt.Println(err)
		http.Error(*w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return true
	}
	return false
}

func badRequest(err error, w *http.ResponseWriter) bool {
	if err != nil {
		fmt.Println(err)
		http.Error(*w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return true
	}
	return false
}
func internalError(err error, w *http.ResponseWriter) bool {
	if err != nil {
		fmt.Println(err)
		http.Error(*w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return true
	}
	return false
}
func newThought(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	reqThought, err := checkThought(r)
	if badRequest(err, &w) {
		return
	}
	bodyReader := bytes.NewReader(reqThought)
	requestUrl := "http://invitation:3002/i/use"
	requestType := "application/json"
	res, err := http.Post(requestUrl, requestType, bodyReader)
	if internalError(err, &w) {
		return
	}
	if res.StatusCode == 403 {
		forbidden(errors.New("invalid code"), &w)
		return
	}

	bodyReader = bytes.NewReader(reqThought)
	requestUrl = "http://azure-thought-storage:3000/createThought"
	http.Post(requestUrl, requestType, bodyReader)

	w.Write([]byte{})
}

func ReverseSlice[T comparable](s []T) {
	sort.SliceStable(s, func(i, j int) bool {
		return i > j
	})
}

func getAllThoughts(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	// Sends request to get all Thoughts
	res, err := http.Get("http://azure-thought-storage:3000/readThoughts")
	if internalError(err, &w) {
		return
	}

	// Get body of response
	body, err := io.ReadAll(res.Body)
	if internalError(err, &w) {
		return
	}
	var resThought *[]Thought
	err = json.Unmarshal(body, &resThought)
	if internalError(err, &w) {
		return
	}
	ReverseSlice(*resThought)
	response, err := json.Marshal(resThought)
	if internalError(err, &w) {
		return
	}

	w.Write(response)
}

// Reads body from request
// Checks for integrity and returns JSON
func checkThought(r *http.Request) ([]byte, error) {
	// Reads request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return []byte{}, err
	}
	fmt.Println(body)
	var reqThought *NewThought
	err = json.Unmarshal(body, &reqThought)
	if err != nil {
		return []byte{}, err
	}

	// Checks Integrity
	if reqThought.History == "" ||
		reqThought.Title == "" ||
		reqThought.Preview == "" {
		fmt.Printf("\nHistory: %v\nTitle: %v\n Preview: %v\n Id: %v\n",
			reqThought.History, reqThought.Title, reqThought.Preview, reqThought.Id)
		return []byte{}, errors.New("invalid request")
	}

	// turns it back to JSON
	response, err := json.Marshal(reqThought)
	if err != nil {
		return []byte{}, err
	}
	return response, nil
}

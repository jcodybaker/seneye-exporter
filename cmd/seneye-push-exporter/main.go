package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jcodybaker/seneye-push-exporter/pkg/lde"
)

func main() {
	var handler http.HandlerFunc = handle
	s := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	msg := []byte("eyJhbGciOiJIUzI1NiJ9.eyJ2ZXJzaW9uIjoiMS4wLjAiLCJTVUQiOnsiaWQiOiJBQUFBQUJCQkJCQ0NDQ0NEREREREVFRUVFRkZGRkYwMCIsIm5hbWUiOiJPZmZpY2UiLCJ0eXBlIjoxLCJUUyI6MTYwOTU2MTIyMiwiZGF0YSI6eyJTIjp7IlciOjEsIlQiOjAsIlAiOjAsIk4iOjAsIlMiOjB9LCJUIjoyMS4xMjUsIlAiOjcuOTQsIk4iOjAuMDAxfX19.PHNLWLYWY7t4c9MYY--KPGxoHkIFgapk0gnPa9wJH2Q")

	request, err := lde.FromRequestBody(msg, []byte("AAAAAAAA"))
	if err != nil {
		fmt.Println("Parsing Body: " + err.Error())
		os.Exit(1)
	}
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("  ", "  ")
	err = out.Encode(request)
	if err != nil {
		fmt.Println("Encoding: " + err.Error())
		os.Exit(1)
	}

	if err := s.ListenAndServe(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.RequestURI)
	for k, values := range r.Header {
		for _, v := range values {
			fmt.Printf("%s: %s\n", k, v)
		}
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(string(b))
	w.WriteHeader(http.StatusOK)
}

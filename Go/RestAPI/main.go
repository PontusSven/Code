package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"

)

type Article struct {
	Title string `json:"title"`
	Desc string `json:"desc"`
	Content string `json:"content"`
}

type Articles []Article

func allArticles(w http.ResponseWriter, r *http.Request) {

	articles := Articles {
		Article{
			Title: "Test Title",
			Desc: "Test Description",
			Content: "Test Content"},
			{
			Title: "Test2 Title",
			Desc: "Test2 Description",
			Content: "Test2 Content"},
		}

	fmt.Println("Endpoint Hit: All Articles Endpoint")
	json.NewEncoder(w).Encode(articles)

}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint Hit")
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/articles", allArticles)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	handleRequests()
}
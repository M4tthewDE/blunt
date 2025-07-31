package main

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/m4tthewde/blunt/components"
	"github.com/m4tthewde/blunt/tmdb"
)

func main() {
	http.Handle("/", templ.Handler(components.Index()))
	http.HandleFunc("/movie_search", movie_search)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}

func movie_search(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	movieSearch := r.FormValue("movie_search")

	resp, err := tmdb.SearchMovies(r.Context(), movieSearch)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	components.MovieSearch(resp.Results).Render(r.Context(), w)
}

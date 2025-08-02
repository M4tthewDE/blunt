package main

import (
	"cmp"
	"log"
	"net/http"
	"slices"

	"github.com/a-h/templ"
	"github.com/m4tthewde/blunt/components"
	"github.com/m4tthewde/blunt/tmdb"
)

func main() {
	http.Handle("/", templ.Handler(components.Index()))
	http.HandleFunc("/movie_search", movie_search)
	http.HandleFunc("GET /movie/{id}", movie)

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

	slices.SortFunc(resp.Results,
		func(a, b tmdb.MovieSearchResult) int {
			return cmp.Compare(b.Popularity, a.Popularity)
		},
	)

	components.MovieSearch(resp.Results).Render(r.Context(), w)
}

func movie(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	movieDetails, err := tmdb.MovieDetails(r.Context(), idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	credits, err := tmdb.Credits(r.Context(), idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	components.Movie(*movieDetails, credits.Cast).Render(r.Context(), w)
}

package main

import (
	"cmp"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/a-h/templ"
	"github.com/m4tthewde/blunt/components"
	"github.com/m4tthewde/blunt/tmdb"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token string `yaml:"token"`
}

var config Config

func main() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", templ.Handler(components.Index()))
	http.HandleFunc("/search", search)
	http.HandleFunc("GET /movie/{id}", movie)
	http.HandleFunc("GET /castMember/{id}", castMember)

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", nil)
}

func search(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	search := r.FormValue("search")

	movieResponse, err := tmdb.SearchMovies(r.Context(), config.Token, search)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	peopleResponse, err := tmdb.SearchPeople(r.Context(), config.Token, search)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	searchResults := make([]components.SearchResult, 0)

	for _, movieResult := range movieResponse.Results {
		searchResults = append(searchResults, components.SearchResult{
			Href:       fmt.Sprintf("/movie/%d", movieResult.Id),
			ImagePath:  tmdb.BuildPosterPath(movieResult.PosterPath),
			Name:       movieResult.OriginalTitle,
			Year:       tmdb.GetReleaseYear(movieResult.ReleaseDate),
			Popularity: movieResult.Popularity,
		})
	}

	for _, peopleResult := range peopleResponse.Results {
		searchResults = append(searchResults, components.SearchResult{
			Href:       fmt.Sprintf("/castMember/%d", peopleResult.Id),
			ImagePath:  tmdb.BuildPosterPath(peopleResult.ProfilePath),
			Name:       peopleResult.Name,
			Year:       "",
			Popularity: peopleResult.Popularity,
		})
	}

	slices.SortFunc(searchResults,
		func(a, b components.SearchResult) int {
			return cmp.Compare(b.Popularity, a.Popularity)
		},
	)

	components.Search(searchResults).Render(r.Context(), w)
}

func movie(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	movieDetails, err := tmdb.MovieDetails(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	credits, err := tmdb.Credits(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	components.Movie(*movieDetails, credits.Cast).Render(r.Context(), w)
}

func castMember(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	people, err := tmdb.People(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	peopleCredits, err := tmdb.PeopleCredits(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	cast := make([]tmdb.PeopleCredit, 0)
	for _, c := range peopleCredits.Cast {
		if c.ReleaseDate != "" {
			cast = append(cast, c)
		}
	}

	slices.SortFunc(cast,
		func(a, b tmdb.PeopleCredit) int {
			timeA, err := time.Parse(time.DateOnly, a.ReleaseDate)
			if err != nil {
				return 0
			}
			timeB, err := time.Parse(time.DateOnly, b.ReleaseDate)
			if err != nil {
				return 0
			}

			if timeA.Before(timeB) {
				return 1
			}

			return -1
		},
	)

	components.CastMember(*people, cast).Render(r.Context(), w)
}

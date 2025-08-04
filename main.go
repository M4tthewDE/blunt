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
	http.HandleFunc("GET /castMember/{id}/graph", castMemberGraph)
	http.HandleFunc("GET /movie/{id}/graph", movieGraph)
	http.HandleFunc("POST /subGraph/movie/{id}", subGraphMovie)
	http.HandleFunc("POST /subGraph/person/{id}", subGraphPerson)

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

func castMemberGraph(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	person, err := tmdb.People(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	credits, err := tmdb.PeopleCredits(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	slices.SortFunc(credits.Cast,
		func(a, b tmdb.PeopleCredit) int {
			return cmp.Compare(b.Popularity, a.Popularity)
		},
	)

	children := make([]components.GraphElement, 0)

	for _, credit := range credits.Cast[0:5] {
		graphElement := components.GraphElement{
			Id:        credit.Id,
			ImagePath: tmdb.BuildPosterPath(credit.PosterPath),
		}

		children = append(children, graphElement)
	}

	parent := components.GraphElement{
		Id:        person.Id,
		ImagePath: tmdb.BuildPosterPath(person.ProfilePath),
	}

	components.Graph(parent, children, "person").Render(r.Context(), w)
}

func movieGraph(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	movie, err := tmdb.MovieDetails(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	credits, err := tmdb.Credits(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	children := make([]components.GraphElement, 0)

	for _, credit := range credits.Cast[0:5] {
		graphElement := components.GraphElement{
			Id:        credit.Id,
			ImagePath: tmdb.BuildPosterPath(credit.ProfilePath),
		}

		children = append(children, graphElement)
	}

	parent := components.GraphElement{
		Id:        movie.Id,
		ImagePath: tmdb.BuildPosterPath(movie.PosterPath),
	}

	components.Graph(parent, children, "movie").Render(r.Context(), w)
}

func subGraphMovie(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	credits, err := tmdb.Credits(r.Context(), config.Token, id)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	children := make([]components.GraphElement, 0)

	for _, credit := range credits.Cast[:5] {
		graphElement := components.GraphElement{
			Id:        credit.Id,
			ImagePath: tmdb.BuildPosterPath(credit.ProfilePath),
		}

		children = append(children, graphElement)
	}

	components.SubGraph(children, "person", credits.Id).Render(r.Context(), w)
}

func subGraphPerson(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	credits, err := tmdb.PeopleCredits(r.Context(), config.Token, id)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	slices.SortFunc(credits.Cast,
		func(a, b tmdb.PeopleCredit) int {
			return cmp.Compare(b.Popularity, a.Popularity)
		},
	)

	children := make([]components.GraphElement, 0)

	for _, credit := range credits.Cast[:5] {
		graphElement := components.GraphElement{
			Id:        credit.Id,
			ImagePath: tmdb.BuildPosterPath(credit.PosterPath),
		}

		children = append(children, graphElement)
	}

	components.SubGraph(children, "movie", credits.Id).Render(r.Context(), w)
}

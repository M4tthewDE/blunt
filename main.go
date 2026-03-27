package main

import (
	"cmp"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
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

//go:embed components/static/*
var static embed.FS

func main() {
	configPath := os.Args[1]
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", templ.Handler(components.Index()))
	http.HandleFunc("/search", search)
	http.HandleFunc("GET /about", about)
	http.HandleFunc("GET /movie/{id}", movie)
	http.HandleFunc("GET /person/{id}", person)
	http.HandleFunc("GET /person/{id}/filteredCastCredits", filteredCastCredits)

	staticFiles, err := fs.Sub(static, "components/static")
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

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
			Href:       fmt.Sprintf("/person/%d", peopleResult.Id),
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

func about(w http.ResponseWriter, r *http.Request) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		log.Println("no build information available")
		w.WriteHeader(500)
		return
	}

	var commitHash string
	var buildTime string
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			commitHash = setting.Value
		}

		if setting.Key == "vcs.time" {
			buildTime = setting.Value
		}
	}

	components.About(commitHash, buildTime).Render(r.Context(), w)
}

func movie(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	movieDetails, err := tmdb.MovieDetails(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	alternativeTitles, err := tmdb.AlternativeTitles(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	credits, err := tmdb.Credits(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	crewMap := make(map[int64]*components.MovieCrewMember)
	for _, c := range credits.Crew {
		if existing, ok := crewMap[c.Id]; ok {
			if !slices.Contains(existing.Jobs, c.Job) {
				existing.Jobs = append(existing.Jobs, c.Job)
			}
		} else {
			crewMap[c.Id] = &components.MovieCrewMember{
				Id:          c.Id,
				ProfilePath: c.ProfilePath,
				Name:        c.Name,
				Jobs:        []string{c.Job},
				Popularity:  c.Popularity,
			}
		}
	}

	crew := make([]components.MovieCrewMember, 0, len(crewMap))
	for _, c := range crewMap {
		crew = append(crew, *c)
	}

	slices.SortFunc(crew, func(a, b components.MovieCrewMember) int {
		return cmp.Compare(b.Popularity, a.Popularity)
	})

	components.Movie(*movieDetails, credits.Cast, crew, alternativeTitles.Titles).Render(r.Context(), w)
}

func person(w http.ResponseWriter, r *http.Request) {
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

	castCreditsFilterParameters := createCastCreditsFilterParameters(peopleCredits.Cast)

	cast := make([]tmdb.CastCredit, 0)
	for _, c := range peopleCredits.Cast {
		if c.ReleaseDate != "" && c.Popularity > castCreditsFilterParameters.Popularity {
			cast = append(cast, c)
		}
	}

	crewMap := make(map[int64]*components.CrewCredit)
	for _, c := range peopleCredits.Crew {
		if c.ReleaseDate == "" {
			continue
		}

		if existing, ok := crewMap[c.Id]; ok {
			if !slices.Contains(existing.Jobs, c.Job) {
				existing.Jobs = append(existing.Jobs, c.Job)
			}
		} else {
			crewMap[c.Id] = &components.CrewCredit{
				Id:            c.Id,
				OriginalTitle: c.OriginalTitle,
				PosterPath:    c.PosterPath,
				ReleaseDate:   c.ReleaseDate,
				Popularity:    c.Popularity,
				Jobs:          []string{c.Job},
			}
		}
	}

	crew := make([]components.CrewCredit, 0, len(crewMap))
	for _, c := range crewMap {
		crew = append(crew, *c)
	}

	slices.SortFunc(cast,
		func(a, b tmdb.CastCredit) int {
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

	slices.SortFunc(crew,
		func(a, b components.CrewCredit) int {
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

	components.Person(*people, cast, crew, castCreditsFilterParameters).Render(r.Context(), w)
}

func createCastCreditsFilterParameters(castCredits []tmdb.CastCredit) components.CastCreditsFilterParameters {
	maxPopularity := 0.0

	for _, castCredit := range castCredits {
		if castCredit.Popularity > maxPopularity {
			maxPopularity = castCredit.Popularity
		}
	}

	minPopularity := maxPopularity

	for _, castCredit := range castCredits {
		if castCredit.Popularity < minPopularity {
			minPopularity = castCredit.Popularity
		}
	}

	return components.CastCreditsFilterParameters{
		Popularity:    2.0,
		MinPopularity: minPopularity,
		MaxPopularity: maxPopularity,
	}
}

func filteredCastCredits(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	popularity := 0.0
	popularityParam := r.URL.Query().Get("popularity")
	if popularityParam != "" {
		popularityVal, err := strconv.ParseFloat(popularityParam, 64)
		if err == nil {
			popularity = popularityVal
		}
	}

	peopleCredits, err := tmdb.PeopleCredits(r.Context(), config.Token, idString)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	filterParams := createCastCreditsFilterParameters(peopleCredits.Cast)
	filterParams.Popularity = popularity

	cast := make([]tmdb.CastCredit, 0)
	for _, c := range peopleCredits.Cast {
		if c.ReleaseDate != "" && c.Popularity >= popularity {
			cast = append(cast, c)
		}
	}

	slices.SortFunc(cast,
		func(a, b tmdb.CastCredit) int {
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

	components.CastCredits(cast, filterParams, idString).Render(r.Context(), w)
}

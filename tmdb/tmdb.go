package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type MovieSearchResult struct {
	OriginalTitle string  `json:"original_title"`
	PosterPath    string  `json:"poster_path"`
	Popularity    float64 `json:"popularity"`
	ReleaseDate   string  `json:"release_date"`
	Id            int64   `json:"id"`
}

type MovieSearchResponse struct {
	Results []MovieSearchResult `json:"results"`
}

type MovieDetailsResponse struct {
	Id               int64   `json:"id"`
	OriginalTitle    string  `json:"original_title"`
	PosterPath       string  `json:"poster_path"`
	Popularity       float64 `json:"popularity"`
	ReleaseDate      string  `json:"release_date"`
	Tagline          string  `json:"tagline"`
	Runtime          int64   `json:"runtime"`
	OriginalLanguage string  `json:"original_language"`
	Overview         string  `json:"overview"`
	Revenue          int64   `json:"revenue"`
}

type MovieCreditsResponse struct {
	Cast []MovieCastMember `json:"cast"`
}

type MovieCastMember struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
}

type PeopleResponse struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
	Birthday    string `json:"birthday"`
	Deathday    string `json:"deathday"`
	Biography   string `json:"biography"`
}

type PeopleCreditsResponse struct {
	Cast []PeopleCredit `json:"cast"`
}

type PeopleCredit struct {
	Id            int64  `json:"id"`
	OriginalTitle string `json:"original_title"`
	PosterPath    string `json:"poster_path"`
	ReleaseDate   string `json:"release_date"`
}

func SearchMovies(ctx context.Context, search string) (*MovieSearchResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.themoviedb.org/3/search/movie", nil)
	if err != nil {
		return nil, err
	}

	token, ok := os.LookupEnv("TMDB_TOKEN")
	if !ok {
		return nil, errors.New("TMDB_TOKEN is not set")
	}

	req.Header.Add("Authorization", "Bearer "+token)

	q := req.URL.Query()

	q.Add("include_adult", "true")
	q.Add("language", "en-US")
	q.Add("page", "1")
	q.Add("query", search)

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response MovieSearchResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func MovieDetails(ctx context.Context, movieId string) (*MovieDetailsResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/movie/%s", movieId), nil)
	if err != nil {
		return nil, err
	}

	token, ok := os.LookupEnv("TMDB_TOKEN")
	if !ok {
		return nil, errors.New("TMDB_TOKEN is not set")
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response MovieDetailsResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func Credits(ctx context.Context, movieId string) (*MovieCreditsResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/credits", movieId), nil)
	if err != nil {
		return nil, err
	}

	token, ok := os.LookupEnv("TMDB_TOKEN")
	if !ok {
		return nil, errors.New("TMDB_TOKEN is not set")
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response MovieCreditsResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func People(ctx context.Context, personId string) (*PeopleResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/person/%s", personId), nil)
	if err != nil {
		return nil, err
	}

	token, ok := os.LookupEnv("TMDB_TOKEN")
	if !ok {
		return nil, errors.New("TMDB_TOKEN is not set")
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response PeopleResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func PeopleCredits(ctx context.Context, personId string) (*PeopleCreditsResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/person/%s/movie_credits", personId), nil)
	if err != nil {
		return nil, err
	}

	token, ok := os.LookupEnv("TMDB_TOKEN")
	if !ok {
		return nil, errors.New("TMDB_TOKEN is not set")
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response PeopleCreditsResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func BuildPosterPath(posterPath string) string {
	return "https://image.tmdb.org/t/p/w600_and_h900_bestv2" + posterPath
}

func GetReleaseYear(releaseDate string) string {
	if releaseDate == "" {
		return "unknown"
	}

	return strings.Split(releaseDate, "-")[0]
}

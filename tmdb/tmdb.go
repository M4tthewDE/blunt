package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	OriginalTitle string  `json:"original_title"`
	PosterPath    string  `json:"poster_path"`
	Popularity    float64 `json:"popularity"`
	ReleaseDate   string  `json:"release_date"`
	Id            int64   `json:"id"`
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

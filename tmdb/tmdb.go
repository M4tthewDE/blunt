package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type PeopleSearchResult struct {
	Name        string  `json:"name"`
	ProfilePath string  `json:"profile_path"`
	Popularity  float64 `json:"popularity"`
	Id          int64   `json:"id"`
}

type PeopleSearchResponse struct {
	Results []PeopleSearchResult `json:"results"`
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
	Id   int64             `json:"id"`
}

type MovieCastMember struct {
	Id          int64   `json:"id"`
	Name        string  `json:"name"`
	Character   string  `json:"character"`
	ProfilePath string  `json:"profile_path"`
	Popularity  float64 `json:"popularity"`
}

type PeopleResponse struct {
	Id                 int64  `json:"id"`
	Name               string `json:"name"`
	ProfilePath        string `json:"profile_path"`
	Birthday           string `json:"birthday"`
	Deathday           string `json:"deathday"`
	Biography          string `json:"biography"`
	KnownForDepartment string `json:"known_for_department"`
	Homepage           string `json:"homepage"`
	PlaceOfBirth       string `json:"place_of_birth"`
}

type PeopleCreditsResponse struct {
	Cast []PeopleCredit `json:"cast"`
	Id   int64          `json:"id"`
}

type PeopleCredit struct {
	Id            int64   `json:"id"`
	OriginalTitle string  `json:"original_title"`
	PosterPath    string  `json:"poster_path"`
	ReleaseDate   string  `json:"release_date"`
	Popularity    float64 `json:"popularity"`
}

func SearchMovies(ctx context.Context, token string, search string) (*MovieSearchResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.themoviedb.org/3/search/movie", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	q := req.URL.Query()

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

func SearchPeople(ctx context.Context, token string, search string) (*PeopleSearchResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.themoviedb.org/3/search/person", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	q := req.URL.Query()

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

	var response PeopleSearchResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func MovieDetails(ctx context.Context, token string, movieId string) (*MovieDetailsResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/movie/%s", movieId), nil)
	if err != nil {
		return nil, err
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

func Credits(ctx context.Context, token string, movieId string) (*MovieCreditsResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/movie/%s/credits", movieId), nil)
	if err != nil {
		return nil, err
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

func People(ctx context.Context, token string, personId string) (*PeopleResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/person/%s", personId), nil)
	if err != nil {
		return nil, err
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

func PeopleCredits(ctx context.Context, token string, personId string) (*PeopleCreditsResponse, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.themoviedb.org/3/person/%s/movie_credits", personId), nil)
	if err != nil {
		return nil, err
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
	if posterPath == "" {
		return "https://www.themoviedb.org/assets/2/v4/glyphicons/basic/glyphicons-basic-38-picture-grey-c2ebdbb057f2a7614185931650f8cee23fa137b93812ccb132b9df511df1cfac.svg"
	}
	return "https://image.tmdb.org/t/p/w600_and_h900_bestv2" + posterPath
}

func GetReleaseYear(releaseDate string) string {
	if releaseDate == "" {
		return ""
	}

	return strings.Split(releaseDate, "-")[0]
}

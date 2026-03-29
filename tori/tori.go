// Package tori contains code for obtaining listing information from tori.fi
package tori

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ToriClient struct {
	client  *http.Client
	baseURL string
}

const ToriQueryURLPath = "/recommerce/forsale/search/api/search/SEARCH_ID_BAP_COMMON"

var (
	ErrEmptyQuery              = errors.New("cannot make request with emtpy query")
	ErrParseListingQuery       = errors.New("failed to parse the listing query response")
	ErrUnsuccessfulHTTPRequest = errors.New("http request returned status code that is not 200")
)

type ToriQueryListingReponse struct {
	Listings ToriQueryListings `json:"docs"`
}

type ToriQueryListing struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Title       string         `json:"heading"`
	Location    string         `json:"location"`
	Image       ToriQueryImage `json:"image"`
	Price       ToriQueryPrice `json:"price"`
	URL         string         `json:"canonical_url"`
	Timestamp   int64          `json:"timestamp"`
	Coordinates Coordinates    `json:"coordinates"`
}

type ToriQueryListings []ToriQueryListing

func (t ToriQueryListings) Len() int           { return len(t) }
func (t ToriQueryListings) Less(i, j int) bool { return t[i].Timestamp < t[j].Timestamp }
func (t ToriQueryListings) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func (t ToriQueryListings) MeanPrice() float64 {
	totalPrice := 0.0

	for _, l := range t {
		totalPrice += float64(l.Price.Amount)
	}

	return totalPrice / float64(len(t))
}

func (t ToriQueryListings) MedianPrice() float64 {
	if len(t)%2 == 1 {
		return float64(t[len(t)/2].Price.Amount)
	} else {
		return (float64(t[len(t)/2].Price.Amount) + float64(t[len(t)/2-1].Price.Amount)) / 2
	}
}

type ToriQueryImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type ToriQueryPrice struct {
	Amount       int    `json:"amount"`
	CurrencyCode string `json:"currency_code"`
	Unit         string `json:"price_unit"`
}

type Coordinates struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

func CreateToriClient(timeout time.Duration) *ToriClient {
	tClient := ToriClient{
		&http.Client{Timeout: timeout},
		"https://tori.fi/",
	}

	return &tClient
}

func (t *ToriClient) GetQueryListings(query string) (ToriQueryListings, error) {
	res, err := t.toriGetQueryListings(query)
	if err != nil {
		return nil, err
	}

	return ParseQueryListings([]byte(res))
}

func (t *ToriClient) toriGetQueryListings(query string) ([]byte, error) {
	var reqBody []byte

	if len(query) == 0 {
		return reqBody, ErrEmptyQuery
	}

	v := url.Values{}
	v.Set("q", query)
	path := ToriQueryURLPath + "?" + v.Encode()

	return t.toriGetRequest(path)
}

func (t *ToriClient) toriGetRequest(path string) ([]byte, error) {
	url := t.baseURL + path

	var reqBody []byte

	res, err := t.client.Get(url)
	if err != nil {
		return reqBody, err
	}

	reqBody, err = io.ReadAll(res.Body)
	if err != nil {
		return reqBody, err
	}

	if res.StatusCode != 200 {
		err = ErrUnsuccessfulHTTPRequest
	}

	return reqBody, err
}

func ParseQueryListings(rBody []byte) (ToriQueryListings, error) {
	listingRes := ToriQueryListingReponse{}

	err := json.Unmarshal(rBody, &listingRes)
	if err != nil {
		return listingRes.Listings, ErrParseListingQuery
	}

	validListings := ToriQueryListings{}
	for _, l := range listingRes.Listings {
		if isValidListing(l) {
			validListings = append(validListings, l)
		}
	}

	return validListings, err
}

func isValidListing(l ToriQueryListing) bool {
	return l.ID != "" && l.Price.Amount != 0
}

package tori

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"
	"time"
)

func TestGetQuery(t *testing.T) {
	t.Run("returns valid data", func(t *testing.T) {
		expected := []byte("test data")
		httpServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(expected)
		}))
		defer httpServ.Close()

		tClient := ToriClient{&http.Client{}, httpServ.URL}
		got, err := tClient.toriGetRequest(ToriQueryURLPath)
		if err != nil {
			t.Fatalf("unexpected error %q", err)
		}

		if !slices.Equal(got, expected) {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("handles invalid response", func(t *testing.T) {
		expected := []byte("bad response")
		httpServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(400)
			w.Write(expected)
		}))
		defer httpServ.Close()

		tClient := ToriClient{&http.Client{}, httpServ.URL}
		got, err := tClient.toriGetRequest("/test")
		if err != ErrUnsuccessfulHTTPRequest {
			t.Errorf("expected err, did not get")
		}
		if !slices.Equal(got, expected) {
			t.Errorf("got %q, expected %q", got, expected)
		}
	})

	t.Run("handles timeout", func(t *testing.T) {
		notExpect := []byte("good")
		httpServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Write(notExpect)
		}))
		defer httpServ.Close()

		tClient := ToriClient{&http.Client{Timeout: 50 * time.Millisecond}, httpServ.URL}
		res, err := tClient.toriGetRequest("/")

		if slices.Equal(res, notExpect) {
			t.Errorf("response should not have completed")
		}
		if err == nil {
			t.Errorf("response should have error")
		}
	})
}

func TestGetListings(t *testing.T) {
	t.Run("requires path", func(t *testing.T) {
		tClient := ToriClient{&http.Client{}, "localhost"}
		_, err := tClient.toriGetQueryListings("")

		if err != ErrEmptyQuery {
			t.Errorf("response should have errored")
		}
	})

	t.Run("query is processed", func(t *testing.T) {
		httpServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("q")
			fmt.Fprint(w, q)
		}))
		defer httpServ.Close()

		expect := "test"

		tClient := ToriClient{&http.Client{}, httpServ.URL}
		got, err := tClient.toriGetQueryListings(expect)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		if !slices.Equal(got, []byte(expect)) {
			t.Errorf("got %q, expected %q", got, expect)
		}
	})
}

func TestParseListings(t *testing.T) {
	singleListing, err := os.ReadFile("../tests/querysingle-valid.json")
	if err != nil {
		panic(err)
	}

	manyListings, err := os.ReadFile("../tests/querylisting-valid.json")
	if err != nil {
		panic(err)
	}

	invalidListing, err := os.ReadFile("../tests/querysingle-bad.json")
	if err != nil {
		panic(err)
	}

	t.Run("single listing", func(t *testing.T) {
		res, err := ParseQueryListings(singleListing)
		if err != nil {
			t.Fatalf("got error %q, did not expect", err)
		}

		if len(res) != 1 || res[0].ID != "39049579" {
			t.Errorf("got %#v, did not expect", res)
		}
	})

	t.Run("many listings", func(t *testing.T) {
		res, err := ParseQueryListings(manyListings)
		if err != nil {
			t.Fatalf("got error %q, did not expect", err)
		}

		expectedCount := 22

		if len(res) != expectedCount {
			t.Errorf("got %d listings, expected %d", len(res), expectedCount)
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		res, err := ParseQueryListings(invalidListing)
		if err != ErrParseListingQuery {
			t.Errorf("got error %q, expected ErrParseListingQuery", err)
		}
		if len(res) != 0 {
			t.Errorf("response should not have returned anything")
		}
	})
}

func TestGetQueryListings(t *testing.T) {
	manyListings, err := os.ReadFile("../tests/querylisting-valid.json")
	if err != nil {
		panic(err)
	}

	invalidListing, err := os.ReadFile("../tests/querysingle-bad.json")
	if err != nil {
		panic(err)
	}

	t.Run("valid response", func(t *testing.T) {
		httpServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(manyListings)
		}))
		defer httpServ.Close()

		tClient := ToriClient{&http.Client{}, httpServ.URL}
		listings, err := tClient.GetQueryListings("test")
		if err != nil {
			t.Fatalf("got error, did not expect %q", err)
		}

		if len(listings) != 22 {
			t.Errorf("got %d listings parsed, expected %d", len(listings), 23)
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		httpServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(invalidListing)
		}))
		defer httpServ.Close()

		tClient := ToriClient{&http.Client{}, httpServ.URL}
		listings, err := tClient.GetQueryListings("test")
		if err == nil {
			t.Errorf("should have received error")
		}

		if len(listings) != 0 {
			t.Errorf("got %d listings parsed, expected %d", len(listings), 0)
		}
	})
}

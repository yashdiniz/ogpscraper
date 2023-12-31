package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/yashdiniz/ogpscraper/internal/metaparser"
	"github.com/yashdiniz/ogpscraper/internal/noembed"
	"github.com/yashdiniz/ogpscraper/internal/opengraph"
)

type server struct {
	RedisConfig  *redis.Options
	DisableCache bool
	CacheTTL     int
}

func NewServer(redis_config *redis.Options, disable_cache bool, cache_ttl int) chi.Router {
	r := chi.NewRouter()
	s := server{redis_config, disable_cache, cache_ttl}

	r.Post("/", s.Scrape)
	return r
}

type ScrapeRequest struct {
	URL     string `json:"url"`
	Refresh bool   `json:"forceRefresh"`
	Raw     bool   `json:"raw"`
}

func (s *server) Scrape(w http.ResponseWriter, r *http.Request) {
	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "require `url`, `forceRefresh` & `raw`", http.StatusBadRequest)
		log.Println("s.Scrape:40", err)
		return
	}

	// TODO: check for cache hit (check if cache is disabled first, or if forceRefresh is set)

	w.Header().Add("Content-Type", "application/json")
	// on cache miss, get meta tags
	tags, err := metaparser.GetMetaTags(req.URL)
	if err != nil {
		// check noembed if colly panics and can't get ogp
		if req.Raw {
			http.Error(w, "`raw` is temporarily not supported", http.StatusFailedDependency)
			log.Println("s.Scrape:53", req.URL, err)
			return
		}
		result, err := noembed.GetNoembedData(req.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not get metadata tags for page: %v", err.Error()), http.StatusFailedDependency)
			log.Println("s.Scrape:59", req.URL, err)
			return
		}
		log.Println("returning noembed data instead")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "could not encode result", http.StatusInternalServerError)
			log.Println("s.Scrape:65", err)
			return
		}
		return // return early since we have already sent a response
	}

	// TODO: cache the results (with ttl of course)

	// if client doesn't want raw results
	if !req.Raw {
		result := opengraph.GetOGPResult(tags)
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "could not encode result", http.StatusInternalServerError)
			log.Println("s.Scrape:78", err)
			return
		}
	} else {
		if err := json.NewEncoder(w).Encode(tags); err != nil {
			http.Error(w, "could not encode result", http.StatusInternalServerError)
			log.Println("s.Scrape:84", err)
			return
		}
	}
}

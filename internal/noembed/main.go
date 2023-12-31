package noembed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/yashdiniz/ogpscraper/internal/opengraph"
)

// REFERENCE: https://noembed.com/

// GetNoembedData calls noembed.com to get the required oembed data, and casts it into an opengraph result
func GetNoembedData(u string) (*opengraph.Result, error) {
	f := url.QueryEscape(u) // safely escape the URL for adding to query

	noembed_url := fmt.Sprintf("https://noembed.com/embed?url=%v&format=json", f)
	res, err := http.Get(noembed_url)
	if err != nil {
		return nil, err
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	var result opengraph.Result
	for key, val := range r {
		switch key {
		case "url":
			result.Url = val.(string)
		case "title":
			result.Title = val.(string)
		case "provider_name":
			result.Description = val.(string)
		case "thumbnail_url":
			result.Image = val.(string)
		}
	}
	result.Type = "website"

	return &result, nil
}

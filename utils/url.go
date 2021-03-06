package utils

import "net/url"

// Converts an slice of URLs to an array of strings
func URLsToStrings(urls ...*url.URL) []string {
	result := []string{}
	for _, u := range urls {
		result = append(result, u.String())
	}
	return result
}

// Returns an URL using only the host and scheme of the other
func BaseURL(rawUrl string) (*url.URL, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	return &url.URL{
		Host:   parsedUrl.Host,
		Scheme: parsedUrl.Scheme,
	}, nil
}

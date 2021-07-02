package helper

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

func call(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Get calls the URL with HTTP GET
func Get(url string) ([]byte, error) {
	return call(http.Get(url))
}

// Post calls the URL with HTTP POST
func Post(url string, header http.Header) ([]byte, error) {
	return PostWithData(url, header, "")
}

// PostWithData calls the URL with HTTP POST
func PostWithData(url string, header http.Header, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte{})
	if !strings.EqualFold(data, "") {
		body = bytes.NewBuffer([]byte(data))
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return call(client.Do(req))
}

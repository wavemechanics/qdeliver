package webdav

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/wavemechanics/deliver/store"
)

type Storage struct {
	url      string
	login    string
	password string
	client   *http.Client
}

func New(url, login, password string) (*Storage, error) {
	return &Storage{
		url:      url,
		login:    login,
		password: password,
		client:   &http.Client{},
	}, nil
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", store.ErrEmptyKey
	}

	path := s.url + "/" + url.PathEscape(key) + ".txt"
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	req.SetBasicAuth(s.login, s.password)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", os.ErrNotExist
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	content, err := ioutil.ReadAll(resp.Body)
	return string(content), nil
}

func (s *Storage) Set(ctx context.Context, key, value string) error {
	if key == "" {
		return store.ErrEmptyKey
	}

	path := s.url + "/" + url.PathEscape(key) + ".txt"

	req, err := http.NewRequest(http.MethodPut, path, strings.NewReader(value))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.SetBasicAuth(s.login, s.password)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

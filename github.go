package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var url = "https://api.github.com/orgs/go-ireul/repos"

func init() {
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) > 0 {
		url = fmt.Sprintf("%s?access_token=%s", url, token)
	}
}

// Repo represents a Github repo
type Repo struct {
	Name        string `json:"name"`
	CloneURL    string `json:"clone_url"`
	Description string `json:"description"`
}

var repos = []Repo{}
var reposMutex = &sync.RWMutex{}
var reposTicker = time.Tick(time.Second * 60)

func startReposTicker() {
	go func() {
		updateRepos()
		for {
			<-reposTicker
			updateRepos()
		}
	}()
}

func updateRepos() (err error) {
	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	reposMutex.Lock()
	defer reposMutex.Unlock()

	return json.Unmarshal(body, &repos)
}

func findRepo(name string) *Repo {
	reposMutex.RLock()
	defer reposMutex.RUnlock()
	for _, repo := range repos {
		if repo.Name == name {
			return &repo
		}
	}
	return nil
}

func listRepos() []Repo {
	reposMutex.RLock()
	defer reposMutex.RUnlock()
	ret := make([]Repo, len(repos))
	copy(ret, repos)
	return ret
}

package main

import "html/template"
import "os"
import "log"
import "strings"
import "net/http"
import "fmt"
import "time"
import "sync"
import "io/ioutil"
import "encoding/json"

var tmplIndex, _ = template.New("index").Parse(`
<!DOCTYPE html>
<html>
	<head>
	  <title>IREUL.com</title>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootswatch/3.3.7/darkly/bootstrap.min.css" integrity="sha256-tfn9eK1pJ8CzrxEY/X948VPX9saxc3sNrzhyU5IX+Yg=" crossorigin="anonymous" />
	</head>
	<body>
		<nav class="navbar navbar-default">
			<div class="container-fluid">
				<div class="navbar-header">
					<a class="navbar-brand" href="/">IREUL.com</a>
				</div>
			</div>
		</nav>

		<div class="container">
			<div class="row">
			  <div class="col">
					<div class="page-header">
						<h3>Packages</h3>
					</div>
				</div>
			</div>

		  <div class="row">
			  <div class="col-md-9">
					<table class="table">
						<thead>
							<tr>
								<th>Name</th>
								<th>Import</th>
								<th>URL</th>
							</tr>
						</thead>
						<tbody>
							{{range .}}
								<tr>
								  <td><a href="/{{.Name}}">{{.Name}}</a></td>
									<td><code>import "ireul.com/{{.Name}}"</code></td>
									<td><a href="{{.CloneURL}}">{{.CloneURL}}</code></td>
								</tr>
							{{end}}
						</tbody>
					</table>
				</div>
			</div>
		</div>
	</body>
</html>
`)

var tmplRepo, _ = template.New("repo").Parse(`
<!DOCTYPE html>
<html>
	<head>
	  <title>{{.Name}} - IREUL.com</title>
		<meta name="go-import" content="ireul.com/{{.Name}} git {{.CloneURL}}" />
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootswatch/3.3.7/darkly/bootstrap.min.css" integrity="sha256-tfn9eK1pJ8CzrxEY/X948VPX9saxc3sNrzhyU5IX+Yg=" crossorigin="anonymous" />
	</head>
	<body>
		<nav class="navbar navbar-default">
			<div class="container-fluid">
				<div class="navbar-header">
					<a class="navbar-brand" href="/">IREUL.com</a>
				</div>
			</div>
		</nav>

	  <div class="container">
		  <div class="row">
			  <div class="col">
					<div class="page-header">
						<h3>{{.Name}}</h3>
						<h4 class="text-muted">{{.Description}}</h4>
					</div>
				</div>
			</div>
			<div class="row">
			  <div class="col">
					<label>Import:</label>
					<p><code>import "ireul.com/{{.Name}}"</code></p>
					<label>URL:</label>
					<p><a href="{{.CloneURL}}">{{.CloneURL}}</a></p>
				</div>
			</div>
		</div>
	</body>
</html>
`)

var url = fmt.Sprintf("https://api.github.com/orgs/go-ireul/repos?access_token=%s", os.Getenv("GITHUB_TOKEN"))

var repos = []Repo{}
var reposMutex = &sync.RWMutex{}
var reposTicker = time.Tick(time.Second * 60)

// Repo represents a Github repo
type Repo struct {
	Name        string `json:"name"`
	CloneURL    string `json:"clone_url"`
	Description string `json:"description"`
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

func main() {
	go func() {
		updateRepos()
		for {
			<-reposTicker
			updateRepos()
		}
	}()

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			res.Header().Set("Content-Type", "text/html")
			tmplIndex.Execute(res, listRepos())
		} else {
			var p *Repo

			components := strings.Split(req.URL.Path, "/")
			if len(components) > 1 {
				p = findRepo(components[1])
			}

			if p != nil {
				res.Header().Set("Content-Type", "text/html")
				tmplRepo.Execute(res, p)
			} else {
				http.NotFound(res, req)
			}
		}
	})

	log.Fatal(http.ListenAndServe(os.Getenv("BIND"), nil))
}

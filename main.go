package main

import (
	"log"
	"os"

	"ireul.com/web"
)

func main() {
	log.SetPrefix("[ireul] ")

	startReposTicker()

	m := web.New()
	m.Use(web.Logger())
	m.Use(web.Logger())
	m.Use(web.Recovery())
	m.Use(web.Static("public"))
	m.Use(web.Renderer())

	m.Get("/", func(ctx *web.Context) {
		ctx.Data["Repos"] = listRepos()
		ctx.HTML(200, "index")
	})

	m.Get("/:name/?*", func(ctx *web.Context) {
		r := findRepo(ctx.Params(":name"))
		if r == nil {
			ctx.HTML(404, "not_found")
		} else {
			ctx.Data["Repo"] = *r
			ctx.HTML(200, "repo")
		}
	})

	m.Run(os.Getenv("HOST"), os.Getenv("PORT"))
}

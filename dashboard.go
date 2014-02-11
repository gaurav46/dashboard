/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"log"
	"net/http"
	"os"
)

var (
	logFlags           = log.Ldate | log.Ltime
	logPrefix          = "[dashboard] "
	productionHostname = "jamesclonk.io"
	currentHostname    = ""
)

type view struct {
	Title string
	Error error
}

func init() {
	log.SetFlags(logFlags)
	log.SetPrefix(logPrefix)

	if hostname, err := hostname(); err != nil {
		log.Fatalf("Encountered a problem while trying to lookup current hostname: %v", err)
	} else {
		currentHostname = hostname.Hostname
	}

	if currentHostname == productionHostname {
		log.Printf("Running on %s, switch to production settings\n", productionHostname)
		martini.Env = martini.Prod
	}
}

func main() {
	m := setupMartini()
	m.Run()
}

func setupMartini() *martini.Martini {
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(martini.Static("assets", martini.StaticOptions{SkipLogging: true})) // skip logging on static content
	m.Use(martini.Logger())
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Layout:     "layout",
		Extensions: []string{".html"},
		Delims:     render.Delims{"{[{", "}]}"},
		IndentJSON: true,
	}))
	m.Map(log.New(os.Stdout, logPrefix, logFlags))
	m.Action(r.Handle)

	setupRoutes(r)

	return m
}

func setupRoutes(r martini.Router) {
	// static
	r.Get("/", func(r render.Render) {
		r.HTML(http.StatusOK, "index", View("Dashboard"))
	})
	r.NotFound(func(r render.Render) {
		r.HTML(http.StatusNotFound, "404", View("404 - Not Found"))
	})

	// api
	r.Get("/api/hostname", DataHandler(hostname()))
	r.Get("/api/ip", DataHandler(ip(currentHostname)))
	r.Get("/api/cpu", DataHandler(cpu()))
	r.Get("/api/mem", DataHandler(mem()))
	r.Get("/api/disk", DataHandler(df()))
}

func DataHandler(data interface{}, err error) func(r render.Render) {
	return func(r render.Render) {
		if err != nil {
			view := View("500 - Internal Server Error")
			view.Error = err
			r.HTML(http.StatusInternalServerError, "500", view)
			return
		}
		r.JSON(http.StatusOK, data)
	}
}

func View(title string) *view {
	return &view{
		Title: title,
	}
}
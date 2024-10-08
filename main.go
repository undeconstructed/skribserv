package main

import (
	"embed"
	"errors"
	"flag"
	"io/fs"
	"net"
	"net/http"
	"os"

	"github.com/undeconstructed/skribserv/app"
	"github.com/undeconstructed/skribserv/lib"
)

//go:embed web/*
var webFiles embed.FS

func makeWebFS(devMode bool) (fs.FS, error) {
	var files fs.FS

	if devMode {
		files = os.DirFS("web")
	} else {
		subFiles, err := fs.Sub(webFiles, "web")
		if err != nil {
			return nil, err
		}

		files = subFiles
	}

	return files, nil
}

func main() {
	devMode := flag.Bool("dev-mode", false, "whether run from source")

	flag.Parse()

	log := lib.MakeLogger(*devMode)

	log.Info("starting", "dev-mode", *devMode)

	files, err := makeWebFS(*devMode)
	if err != nil {
		log.Error("make web fs", "err", err)
		os.Exit(1)
	}

	lr, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Error("bind port", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", lib.Middleware(log, true, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, files, r.URL.Path)
	}))

	theApp, err := app.New()
	if err != nil {
		log.Error("make api", "err", err)
		os.Exit(1)
	}

	theApp.Install(func(pattern string, handler http.HandlerFunc) {
		mux.HandleFunc(pattern, lib.Middleware(log, !*devMode, handler))
	})

	srv := http.Server{
		Handler: mux,
	}

	err = srv.Serve(lr)
	if !errors.Is(err, http.ErrServerClosed) {
		log.Error("server", "err", err)
	}
}

package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/undeconstructed/skribserv/agordoj"
	"github.com/undeconstructed/skribserv/apo"
	"github.com/undeconstructed/skribserv/db"
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

	logLevel := slog.LevelInfo

	if *devMode {
		logLevel = slog.LevelDebug
	}

	if l := os.Getenv("LOG_LEVEL"); l != "" {
		// maybe there is another log level parser, but I can't only find the json one, which needs all this extra rubbish
		err := logLevel.UnmarshalJSON([]byte(strconv.Quote(l)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "bad log level: %v\n", err)
			os.Exit(2)
		}
	}

	slog.SetDefault(lib.MakeLogger(logLevel, true))

	log := lib.DefaultLog(context.Background())

	config, path, err := agordoj.ReadConfig("skribsrv.yaml")
	if err != nil {
		log.Error("read config", "err", err)
		os.Exit(1)
	}

	slog.Info("config", "path", path, "dev-mode", *devMode, "bind", config.ListenAddr)

	// db

	db, err := db.Munti(config.DBDSN, log.Raw().With("so", "db"))
	if err != nil {
		log.Error("connect db", "err", err)
		os.Exit(1)
	}

	// base server

	files, err := makeWebFS(*devMode)
	if err != nil {
		log.Error("make web fs", "err", err)
		os.Exit(1)
	}

	lr, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		log.Error("bind port", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mw := lib.BasicMiddleware(!*devMode)

	mux.HandleFunc("GET /", mw(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, files, r.URL.Path)
	}))

	// app

	laApo, err := apo.Nova(db, log.Raw().With("so", "apo"))
	if err != nil {
		log.Error("make api", "err", err)
		os.Exit(1)
	}

	laApo.Muntiĝi(func(method, path string, handler http.HandlerFunc, mws ...lib.MiddlewareFunc) {
		for _, m := range mws {
			handler = m(handler)
		}
		mux.HandleFunc(method+" "+path, mw(handler))
	})

	// serve

	srv := http.Server{
		Handler: mux,
	}

	err = srv.Serve(lr)
	if !errors.Is(err, http.ErrServerClosed) {
		log.Error("server", "err", err)
	}
}

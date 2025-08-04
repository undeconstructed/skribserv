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

	"github.com/undeconstructed/skribserv/app"
	"github.com/undeconstructed/skribserv/config"
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

	config, path, err := config.ReadConfig("skribsrv.yaml")
	if err != nil {
		log.Error("read config", "err", err)
		os.Exit(1)
	}

	slog.Info("config", "path", path, "dev-mode", *devMode, "bind", config.ListenAddr)

	// db

	db, err := db.Setup(config.DBDSN, log.Raw().With("so", "db"))
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

	// mux.HandleFunc("GET /", mw(func(w http.ResponseWriter, r *http.Request) {
	// 	x := *r.URL
	// 	f := &x
	// 	f.Scheme = "http"
	// 	f.Host = "localhost:5173"

	// 	rf, err := http.Get(f.String())
	// 	if err != nil {
	// 		log.Error("proxy", "err", err)
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		return
	// 	}

	// 	for k, v := range rf.Header {
	// 		w.Header().Add(k, v[0])
	// 	}

	// 	w.WriteHeader(rf.StatusCode)

	// 	body := rf.Body
	// 	defer body.Close()

	// 	io.Copy(w, body)
	// }))

	// app

	theApp, err := app.New(db, log.Raw().With("so", "apo"))
	if err != nil {
		log.Error("make api", "err", err)
		os.Exit(1)
	}

	theApp.Mount(func(method, path string, handler http.HandlerFunc, mws ...lib.MiddlewareFunc) {
		for _, m := range mws {
			handler = m(handler)
		}

		mux.HandleFunc(method+" /api"+path, mw(handler))
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

package main

import (
	"log/slog"
	"net/http"
)

type App struct {
	Server *http.Server
	Log    *slog.Logger
}

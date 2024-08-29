package main

import (
	"fmt"
	"net/http"
)

func setupServer(config *Conf) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.FileServer(http.Dir(config.Root)),
	}
}

package handlers

import "net/http"

func GetFileServerHandler() http.Handler {
	webDir := "./web"

	return http.FileServer(http.Dir(webDir))
}

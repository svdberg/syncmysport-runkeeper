package api

import (
	"log"
	"net/http"
)

type HttpsRedirectingFileHandler interface {
	http.Handler
	fileHandler() http.Handler
}

type httpsHandler struct {
	internalHandler http.Handler
}

func NewHttpsRedirectFileHandler(dir http.Dir) HttpsRedirectingFileHandler {
	handler := &httpsHandler{}
	handler.internalHandler = http.FileServer(dir)
	return handler
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		// see @andreiavrammsd comment: often 307 > 301
		http.StatusTemporaryRedirect)
}

func isXForwardedHTTPS(request *http.Request) bool {
	xForwardedProto := request.Header.Get("X-Forwarded-Proto")

	return len(xForwardedProto) > 0 && xForwardedProto == "https"
}

func (h *httpsHandler) fileHandler() http.Handler {
	return h.internalHandler
}

func (h *httpsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !isXForwardedHTTPS(req) {
		//redirect to https
		redirect(w, req)
	} else {
		h.fileHandler().ServeHTTP(w, req)
	}
}

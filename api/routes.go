package api

import (
	"net/http"

	"github.com/svdberg/syncmysport-runkeeper/Godeps/_workspace/src/github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

var routes = Routes{
	Route{
		"Test",
		"GET",
		"/test",
		TestCookie,
	},
	Route{
		"RKOAuthCallback",
		"GET",
		"/code",
		OAuthCallback,
	},
	Route{
		"SyncTaskIndex",
		"GET",
		"/synctasks",
		SyncTaskIndex,
	},
	Route{
		"SyncTaskShow",
		"GET",
		"/synctasks/{syncTaskId}",
		SyncTaskShow,
	},
	Route{
		"SyncTaskCreate",
		"POST",
		"/synctasks",
		SyncTaskCreate,
	},
}

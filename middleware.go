package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"strings"
)

// Use is a basic middleware chainer.
// This function allows an infinite amount of middleware to be called before the final handler ("h") is called.
func Use(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

// PRTMiddleware is a validation middleware for making sure that the platform, region and tag are all valid.
// In this case, platform and region both have a limited number of possible options the caller can choose from.
// The validation is pushed off to another function that will return a list of error strings in the event validation fails.
func PRTMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the platform, region and tag from the request URL.
		// Pack these into a new Player for future use.
		vars := mux.Vars(r)
		p := getPlayer(vars)

		errors := []string{}
		if !p.platformIsValid() {
			errors = append(errors, ERROR_BAD_PLATFORM)
		}

		if !p.regionIsValid() {
			errors = append(errors, ERROR_BAD_REGION)
		}

		// If there were no errors, the handler in the "chain" will now be called.
		// Otherwise, we need to bail since there are errors preventing this function from completing.
		if len(errors) == 0 {
			h.ServeHTTP(w, r)
		} else {
			MarshalAndHandleErrors(w, r, ErrorResponse{Errors: errors})
		}
	})
}

// PRTMiddleware is a validation middleware for making sure that the platform, region, tag and mode are all valid.
// NOTE: This is a copy of PRTMiddleware, with the addition of mode validation.
func PRTMMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the platform, region and tag from the request URL.
		// Pack these into a new Player for future use.
		vars := mux.Vars(r)
		p := getPlayer(vars)

		errors := []string{}
		if !p.platformIsValid() {
			errors = append(errors, ERROR_BAD_PLATFORM)
		}

		if !p.regionIsValid() {
			errors = append(errors, ERROR_BAD_REGION)
		}

		if !modeIsValid(vars["mode"]) {
			errors = append(errors, ERROR_BAD_MODE)
		}

		// If there were no errors, the handler in the "chain" will now be called.
		// Otherwise, we need to bail since there are errors preventing this function from completing.
		if len(errors) == 0 {
			h.ServeHTTP(w, r)
		} else {
			MarshalAndHandleErrors(w, r, ErrorResponse{Errors: errors})
		}
	})
}

func modeIsValid(mode string) bool {
	if !modes[strings.ToLower(mode)] {
		return false
	}

	return true
}
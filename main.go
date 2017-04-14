package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"encoding/json"
	"os"
	"github.com/gorilla/handlers"
)

func main() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	router := mux.NewRouter().StrictSlash(true)

	// "Root" / "Home" route
	router.HandleFunc("/", home).Methods(http.MethodGet)

	APIRouter := router.PathPrefix("/api").Subrouter()
	//APIRouter.Path("/patch-notes").HandlerFunc(goverwatch.PatchNoteHandler).Methods(http.MethodGet)
	APIRouter.Path("/search/{tag}").HandlerFunc(SearchHandler).Methods(http.MethodGet)

	PRTRouter := APIRouter.PathPrefix("/{platform}/{region}/{tag}").Subrouter()
	PRTRouter.Handle("/profile", Use(http.HandlerFunc(ProfileHandler), PRTMiddleware)).Methods(http.MethodGet)
	PRTRouter.Handle("/achievements", Use(http.HandlerFunc(AchievementsHandler), PRTMiddleware)).Methods(http.MethodGet)

	// Any route under "/{platform}/{region}/{tag}/{mode}"
	PRTMRouter := APIRouter.PathPrefix("/{platform}/{region}/{tag}/{mode}").Subrouter()
	PRTMRouter.Handle("/all-hero-stats", Use(http.HandlerFunc(AllHeroStatsHandler), PRTMMiddleware)).Methods(http.MethodGet)
	PRTMRouter.Handle("/heros-breakdown", Use(http.HandlerFunc(HerosHandler), PRTMMiddleware)).Methods(http.MethodGet)

	// TODO: Hero name validation
	PRTMRouter.Handle("/hero/{name}", Use(http.HandlerFunc(HeroHandler), PRTMMiddleware)).Methods(http.MethodGet)

	log.Println("Listening on " + PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, handlers.CORS()(router)))
}

func home(w http.ResponseWriter, r *http.Request) {
	response, err := json.Marshal("API is online")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

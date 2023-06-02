package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func groupRoutes(router chi.Router) {
	router.With(middlewareAUth).Route("/{username}", func(r chi.Router) {
		r.Post("/add/", addBook)
		r.Delete("/delete/", removeBook)
		r.Route("/read", func(r chi.Router) {
			r.Get("/list/", listBooks)
		})
		r.Route("/unread", func(r chi.Router) {
			r.Get("list/", listBooks)
		})
		r.Get("list_all/", listAllBooks)
	})
}

func main() {

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Group(groupRoutes)
	//router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Post("/register/", CreateAccount)
	router.Post("/login/", authHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
}

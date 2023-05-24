package handler

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"library/internal/handler/rest"
	"library/internal/service"
)

type Dependencies struct {
	AuthorService service.AuthorService
	BookService   service.BookService
	MemberService service.MemberService
}

type Handler struct {
	HTTP *chi.Mux
}

func New(d Dependencies) Handler {
	// Init a new router instance
	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Use(middleware.RealIP)

	r.Use(middleware.Logger)

	r.Use(middleware.Recoverer)

	r.Use(middleware.CleanPath)

	r.Use(middleware.Heartbeat("/ping"))

	r.Use(middleware.Timeout(time.Second * 60))

	r.Use(middleware.AllowContentType("application/json"))

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/authors", rest.NewAuthorHandler(d.AuthorService).Routes())
		r.Mount("/books", rest.NewBookHandler(d.BookService).Routes())
		r.Mount("/members", rest.NewMemberHandler(d.MemberService).Routes())
	})

	return Handler{
		HTTP: r,
	}
}

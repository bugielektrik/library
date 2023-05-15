package rest

import (
	"context"
	"library/config"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"library/internal/service"
)

type Dependencies struct {
	Configs       config.Config
	AuthorService service.AuthorService
	BookService   service.BookService
	MemberService service.MemberService
}

type Server struct {
	Router *chi.Mux
	*http.Server
}

func New(d Dependencies) *Server {
	// Init a new router instance
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(middleware.AllowContentType("application/json"))

	// Health check
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Register a new routes
	router.Mount("/api/author", AuthorRoutes(d.AuthorService))
	router.Mount("/api/book", BookRoutes(d.BookService))
	router.Mount("/api/member", MemberRoutes(d.MemberService))

	return &Server{
		Router: router,
		Server: &http.Server{
			Handler:        router,
			Addr:           ":" + d.Configs.HTTP.Port,
			ReadTimeout:    d.Configs.HTTP.ReadTimeout,
			WriteTimeout:   d.Configs.HTTP.WriteTimeout,
			MaxHeaderBytes: d.Configs.HTTP.MaxHeaderMegabytes << 20,
		},
	}
}

func (s Server) Run() (err error) {
	return s.ListenAndServe()
}

func (s Server) Stop(ctx context.Context) (err error) {
	return s.Shutdown(ctx)
}

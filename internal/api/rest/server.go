package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"library/config"
	"library/internal/service"
)

type Dependencies struct {
	Configs       config.Config
	AuthorService service.AuthorService
	BookService   service.BookService
	MemberService service.MemberService
}

type Server struct {
	*http.Server
}

func New(d Dependencies) *Server {
	// Init a new r instance
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.AllowContentType("application/json"))

	// Health check
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Register a new routes
	r.Mount("/api/author", AuthorRoutes(d.AuthorService))
	r.Mount("/api/book", BookRoutes(d.BookService))
	r.Mount("/api/member", MemberRoutes(d.MemberService))

	return &Server{
		Server: &http.Server{
			Handler:        r,
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

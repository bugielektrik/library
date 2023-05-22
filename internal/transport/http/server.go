package http

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

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

func New(d Dependencies) Server {
	// Init a new r instance
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

	r.Mount("/api/author", NewAuthorHandler(d.AuthorService).Routes())
	r.Mount("/api/book", NewBookHandler(d.BookService).Routes())
	r.Mount("/api/member", NewMemberHandler(d.MemberService).Routes())

	return Server{&http.Server{
		Handler:        r,
		Addr:           ":" + d.Configs.HTTP.Port,
		ReadTimeout:    d.Configs.HTTP.ReadTimeout,
		WriteTimeout:   d.Configs.HTTP.WriteTimeout,
		MaxHeaderBytes: d.Configs.HTTP.MaxHeaderMegabytes << 20,
	}}
}

func (s Server) Run() (err error) {
	if err = s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return
	}

	return
}

func (s Server) Stop(ctx context.Context) (err error) {
	return s.Shutdown(ctx)
}

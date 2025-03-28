package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/ohmpatel1997/findhotel/lib/config"
	zlog "github.com/ohmpatel1997/findhotel/lib/log"
)

// Router interface, a subset of chi with some convenience methods
type Router interface {
	Delete(string, http.HandlerFunc, ...func(http.Handler) http.Handler)
	Get(string, http.HandlerFunc, ...func(http.Handler) http.Handler)
	Patch(string, http.HandlerFunc, ...func(http.Handler) http.Handler)
	Post(string, http.HandlerFunc, ...func(http.Handler) http.Handler)
	Put(string, http.HandlerFunc, ...func(http.Handler) http.Handler)
	Options(string, http.HandlerFunc, ...func(http.Handler) http.Handler)

	Route(string, func(r Router)) Router

	Handle(string, http.Handler)
	HandleFunc(string, http.HandlerFunc)
	With(middlewares ...func(http.Handler) http.Handler) Router

	ListenAndServeTLS(cfg *config.Server) error
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type router struct {
	chi *chi.Mux
}

// NewBasicRouter is a basic router without authorization for back compat
func NewBasicRouter() Router {
	rchi := chi.NewRouter()
	rchi.Use(LoggerAndRecover)

	return &router{
		chi: rchi,
	}
}

func (r *router) With(middlewares ...func(http.Handler) http.Handler) Router {
	r.chi = r.chi.With(middlewares...).(*chi.Mux)
	return r
}

func (r *router) Delete(p string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.chi.With(middlewares...).Delete(p, h)
}

func (r *router) Get(p string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.chi.With(middlewares...).Get(p, h)
}

func (r *router) Patch(p string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.chi.With(middlewares...).Patch(p, h)
}

func (r *router) Post(p string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.chi.With(middlewares...).Post(p, h)
}

func (r *router) Put(p string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.chi.With(middlewares...).Put(p, h)
}

func (r *router) Options(p string, h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) {
	r.chi.With(middlewares...).Options(p, h)
}

func (r *router) Route(p string, fn func(r Router)) Router {
	nr := &router{chi.NewRouter()} //get new router

	if fn != nil {
		fn(nr) //register the sub path
	}

	r.Mount(p, nr) //mount the new router
	return nr
}

func (r *router) Mount(p string, h http.Handler) {
	r.chi.Mount(p, h)
}

func (r *router) Handle(p string, h http.Handler) {
	r.chi.Handle(p, h)
}

func (r *router) HandleFunc(p string, h http.HandlerFunc) {
	r.chi.HandleFunc(p, h)
}

func (r *router) ListenAndServeTLS(cfg *config.Server) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r.chi,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		switch {
		case err != nil && !errors.Is(err, http.ErrServerClosed):
			zlog.Logger().Error("error running server", err, nil)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zlog.Logger().Info("gracefully shutting down the servers...!", nil)

	localCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(localCtx); err != nil {
		zlog.Logger().Error("Forcefully shutting down the server", err, nil)
	}

	zlog.Logger().Info("server shutdown", nil)
	return nil
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}

//Response is all the info we need to properly render json ResponseWriter, Data, Logger, Status
type Response struct {
	Writer http.ResponseWriter
	Data   interface{}
	Status int
}

type HttpError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("%s:%d", e.Message, e.Status)
}

func NewHttpError(message string, status int) *HttpError {
	return &HttpError{message, status}
}

func RenderJSON(r Response) {
	var j []byte
	var err error

	j, err = json.Marshal(r.Data)

	r.Writer.Header().Set("Content-Type", "application/json")

	if err != nil {
		r.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Status > 0 {
		r.Writer.WriteHeader(r.Status)
	}

	r.Writer.Write(j)
}

func RenderError(r http.ResponseWriter, err error) {
	httpErr := &HttpError{
		Message: "Internal Server Error",
		Status:  500,
	}
	r.Header().Set("Content-Type", "application/json")

	var httpError *HttpError
	switch {
	case errors.As(err, &httpError):
		httpErr.Status = httpError.Status
		httpErr.Message = httpError.Message
	}

	r.WriteHeader(httpError.Status)
	resp, err := json.Marshal(httpError)
	if err != nil {
		r.Write([]byte("internal server error"))
		return
	}
	r.Write(resp)
}

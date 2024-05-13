package worker

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

type Api struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *chi.Mux
	Logger  *slog.Logger
}

func tasksRoutes(a *Api) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", a.StartTaskHandler)
		r.Get("/", a.GetTasksHandler)
		r.Route("/{taskID}", func(r chi.Router) {
			r.Delete("/", a.StopTaskHandler)
		})
	}
}

func statsRoutes(a *Api) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", a.GetStatsHandler)
	}
}

func (a *Api) initRouter() {
	a.Router = chi.NewRouter()
	a.Router.Route("/tasks", tasksRoutes(a))
	a.Router.Route("/stats", statsRoutes(a))
}

func (a *Api) Start() {
	a.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	a.initRouter()
	http.ListenAndServe(fmt.Sprintf("%s:%d", a.Address, a.Port), a.Router)
}

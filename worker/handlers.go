package worker

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/michaelgov-ctrl/mellowyellow/task"
)

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	t := task.TaskEvent{}
	if err := d.Decode(&t); err != nil {
		msg := fmt.Errorf("error unmarshalling body: %v", err)
		a.Logger.Error(msg.Error())

		w.WriteHeader(http.StatusBadRequest)
		e := fmt.Sprintf("ErrResponse{ HTTPStatusCode: %d, Message: %s, }", http.StatusBadRequest, msg)
		json.NewEncoder(w).Encode(e)

		return
	}

	a.Worker.AddTask(t.Task)
	a.Logger.Info(fmt.Sprintf("Added task %v\n", t.Task.ID.String()))

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t.Task)
}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a.Worker.ListTasks())
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		msg := fmt.Errorf("empty request body from %v", r.Host)
		a.Logger.Error(msg.Error())

		w.WriteHeader(http.StatusBadRequest)
		e := fmt.Sprintf("ErrResponse{ HTTPStatusCode: %d, Message: %s, }", http.StatusBadRequest, msg)
		json.NewEncoder(w).Encode(e)

		return
	}

	tID, err := uuid.Parse(taskID)
	if err != nil {
		msg := fmt.Errorf("unable to parse request as uuid: %v", err)
		a.Logger.Error(msg.Error())

		w.WriteHeader(http.StatusBadRequest)
		e := fmt.Sprintf("ErrResponse{ HTTPStatusCode: %d, Message: %s, }", http.StatusBadRequest, msg)
		json.NewEncoder(w).Encode(e)

		return
	}

	taskToStop, ok := a.Worker.Db[tID]
	if !ok {
		msg := fmt.Sprintf("no task with ID %v found", tID)
		a.Logger.Error(msg)

		w.WriteHeader(http.StatusBadRequest)
		e := fmt.Sprintf("ErrResponse{ HTTPStatusCode: %d, Message: %s, }", http.StatusBadRequest, msg)
		json.NewEncoder(w).Encode(e)

		return
	}

	cp := *taskToStop
	cp.State = task.Completed

	a.Worker.AddTask(cp)

	a.Logger.Info(fmt.Sprintf("added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerID))

	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.Stats)
}

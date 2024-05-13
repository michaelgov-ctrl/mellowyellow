package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"

	"github.com/michaelgov-ctrl/mellowyellow/task"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
	Stats     *Stats
}

func (w *Worker) CollectStats() {
	for {
		log.Println("collecting stats")
		w.Stats, w.Stats.TaskCount = GetStats(), w.TaskCount
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) RunTask() task.DockerResult {
	var res task.DockerResult

	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("No tasks in the queue")
		res.Error = nil
		return res
	}

	taskQueued := t.(task.Task)

	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = &taskQueued
	}

	if task.ValidStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			res = w.StartTask(taskQueued)
		case task.Completed:
			res = w.StopTask(taskQueued)
		default:
			res.Error = errors.New("we should not get here")
		}
	} else {
		res.Error = fmt.Errorf("invalid transition from %v to %v", taskPersisted.State, taskQueued.State)
	}

	return res
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	cfg := t.NewConfig()
	d := cfg.NewDocker()

	fmt.Printf("%v\n%v\n", cfg, d)
	res := d.Run()
	if res.Error != nil {
		log.Printf("Err running task %v: %v\n", t.ID, res.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return res
	}

	t.ContainerID = res.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t

	return res
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	cfg := t.NewConfig()
	d := cfg.NewDocker()

	res := d.Stop(t.ContainerID)
	if res.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ContainerID, res.Error)
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t

	log.Printf("Stopped and removed container %v for task %v\n", t.ContainerID, t.ID)

	return res
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) ListTasks() []task.Task {
	res := []task.Task{}

	for _, v := range w.Db {
		res = append(res, *v)
	}

	return res
}

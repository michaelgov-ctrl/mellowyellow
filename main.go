package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/michaelgov-ctrl/mellowyellow/task"
	"github.com/michaelgov-ctrl/mellowyellow/worker"
)

func main() {
	host, port := "localhost", 4000

	fmt.Println("Starting cube worker")

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	api := worker.Api{
		Address: host,
		Port:    port,
		Worker:  &w,
	}

	go runTasks(&w)
	go w.CollectStats()
	api.Start()
}

func runTasks(w *worker.Worker) {
	for {
		if w.Queue.Len() != 0 {
			res := w.RunTask()
			if res.Error != nil {
				log.Printf("Error running task: %v\n", res.Error)
			}
		} else {
			log.Println("No tasks to process currently")
		}
		log.Println("Sleeping for 10 seconds.")
		time.Sleep(time.Second * 10)
	}
}

/*
func main() {
	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	t := task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "strm/helloworld-http",
	}

	fmt.Println("starting task")
	w.AddTask(t)
	res := w.RunTask()
	if res.Error != nil {
		panic(res.Error)
	}

	t.ContainerID = res.ContainerId

	fmt.Printf("task %s is runing in container %s\n", t.ID, t.ContainerID)
	fmt.Println("sleepy time")
	time.Sleep(time.Minute * 5)

	fmt.Printf("stopping task %s\n", t.ID)
	t.State = task.Completed
	w.AddTask(t)
	res = w.RunTask()
	if res.Error != nil {
		panic(res.Error)
	}
}
*/

/*

func main() {
	t := task.Task{
		ID:     uuid.New(),
		Name:   "Task-42",
		State:  task.Pending,
		Image:  "Image-42",
		Memory: 1024,
		Disk:   1,
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		Timestamp: time.Now(),
		Task:      t,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("task event: %v\n", te)

	w := worker.Worker{
		Name:  "worker-42",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	fmt.Printf("worker: %v\n", w)
	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StopTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string][]*task.Task),
		EventDb: make(map[string][]*task.TaskEvent),
		Workers: []string{w.Name},
	}
	fmt.Printf("manager: %v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SendWork()

	n := node.Node{
		Name:   "Node-42",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 1024,
		Disk:   25,
		Role:   "worker",
	}
	fmt.Printf("node: %v\n", n)

	fmt.Printf("create a test container\n")
	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Printf("%v", createResult.Error)
		os.Exit(1)
	}
	defer dockerTask.Client.Close()

	time.Sleep(5 * time.Second)

	fmt.Printf("stopping container %s\n", createResult.ContainerId)
	_ = stopContainer(dockerTask, createResult.ContainerId)
}


func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=mellowyellow",
			"POSTGRES_PASSWORD=secret",
		},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Client: dc,
		Config: c,
	}

	result := d.Run()
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, nil
	}

	fmt.Printf("container %s is running with config %v\n", result.ContainerId, c)
	return &d, &result
}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	res := d.Stop(id)
	if res.Error != nil {
		fmt.Printf("%v\n", res.Error)
		return nil
	}

	fmt.Printf("container %s has been stopped and removed\n", res.ContainerId)
	return &res
}
*/

package task

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/moby/moby/pkg/stdcopy"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

var stateTransitionMap = map[State][]State{
	Pending:   {Scheduled},
	Scheduled: {Scheduled, Running, Failed},
	Running:   {Running, Completed, Failed},
	Completed: {},
	Failed:    {},
}

func Contains(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}

	return false
}

func ValidStateTransition(src State, dst State) bool {
	return Contains(stateTransitionMap[src], dst)
}

type Task struct {
	ID            uuid.UUID         `json:"id,omitempty"`
	ContainerID   string            `json:"container_id,omitempty"`
	Name          string            `json:"name,omitempty"`
	State         State             `json:"state,omitempty"`
	Image         string            `json:"image,omitempty"`
	Cpu           float64           `json:"cpu,omitempty"`
	Memory        int               `json:"memory,omitempty"`
	Disk          int               `json:"disk,omitempty"`
	ExposedPorts  nat.PortSet       `json:"exposed_ports,omitempty"`
	PortBindings  map[string]string `json:"port_binding,omitempty"`
	RestartPolicy string            `json:"restart_policy,omitempty"`
	StartTime     time.Time         `json:"start_time,omitempty"`
	FinishTime    time.Time         `json:"finish_time,omitempty"`
}

func (t *Task) NewConfig() Config {
	c := Config{
		Name:  t.Name,
		Image: t.Image,
	}

	return c
}

type TaskEvent struct {
	ID        uuid.UUID `json:"id,omitempty"`
	State     State     `json:"state,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Task      Task      `json:"task,omitempty"`
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy RestartPolicy
	Runtime       Runtime
}

func (c *Config) NewDocker() Docker {
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := Docker{
		Client: dc,
		Config: *c,
	}

	return d
}

type RestartPolicy string

func (rp *RestartPolicy) convert() container.RestartPolicyMode {
	//TODO convert string to container type

	return container.RestartPolicyAlways
}

type Runtime struct {
	ContainerID string
}

type Docker struct {
	Client *client.Client
	Config Config
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func (d *Docker) Run() DockerResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reader, err := d.Client.ImagePull(ctx, d.Config.Image, image.PullOptions{})
	if err != nil {
		log.Printf("error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	io.Copy(os.Stdout, reader)

	rp := container.RestartPolicy{
		Name: d.Config.RestartPolicy.convert(),
	}

	r := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * math.Pow(10, 9)),
	}

	cc := container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(context.TODO(), &cc, &hc, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Error creating container using image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	if err := d.Client.ContainerStart(context.TODO(), resp.ID, container.StartOptions{}); err != nil {
		log.Printf("error starting container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	fmt.Printf("-------------------%s-------------------\n", resp.ID)
	out, err := d.Client.ContainerLogs(context.TODO(), resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Printf("error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{ContainerId: resp.ID, Action: "start", Result: "success"}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("attempting to stop container %v\n", id)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := d.Client.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
		log.Printf("Error stopping container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	opts := container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	}

	if err := d.Client.ContainerRemove(context.TODO(), id, opts); err != nil {
		log.Printf("error removing container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{ContainerId: id, Action: "stop", Result: "success", Error: nil}
}

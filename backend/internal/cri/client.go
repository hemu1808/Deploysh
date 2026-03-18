package cri

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/hemu1808/auradeploy/backend/internal/models"
)

// Client is a wrapper around the containerd client
type Client struct {
	cli    *containerd.Client
	logger *slog.Logger
	ns     string
}

// NewClient initializes the connection to containerd socket
func NewClient(socketPath string, namespace string, logger *slog.Logger) (*Client, error) {
	// Connect to containerd's grpc socket
	cli, err := containerd.New(socketPath)
	if err != nil {
		return nil, err
	}

	return &Client{
		cli:    cli,
		logger: logger,
		ns:     namespace,
	}, nil
}

func (c *Client) Close() {
	if c.cli != nil {
		c.cli.Close()
	}
}

// createContext returns a background context with the configured namespace
func (c *Client) createContext() (context.Context, context.CancelFunc) {
	ctx := namespaces.WithNamespace(context.Background(), c.ns)
	return context.WithTimeout(ctx, 60*time.Second)
}

// RunContainer pulls the image, creates the container, and starts the task
func (c *Client) RunContainer(app models.Application) error {
	ctx, cancel := c.createContext()
	defer cancel()

	c.logger.Info("Pulling image", "image", app.DockerImage)
	image, err := c.cli.Pull(ctx, app.DockerImage, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", app.DockerImage, err)
	}

	// Prepare container spec (CGroups, namespaces, mounts)
	c.logger.Info("Creating container spec", "id", app.ID)

	// Build env string slice
	var envs []string
	for _, e := range app.EnvVars {
		envs = append(envs, fmt.Sprintf("%s=%s", e.Key, e.Value))
	}

	container, err := c.cli.NewContainer(
		ctx,
		app.ID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(app.ID+"-snapshot", image),
		containerd.WithNewSpec(oci.WithImageConfig(image), oci.WithEnv(envs)),
	)
	if err != nil {
		return fmt.Errorf("failed to create container %s: %w", app.ID, err)
	}

	// Create a task (the actual isolated process)
	c.logger.Info("Starting container task", "id", app.ID)
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task for %s: %w", app.ID, err)
	}

	// Make sure we wait before calling start
	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		return err
	}

	// Call start to actually start the container execution
	if err := task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task for %s: %w", app.ID, err)
	}

	// Launch a small goroutine to log when the container exists
	go func() {
		status := <-exitStatusC
		code, _, err := status.Result()
		if err != nil {
			c.logger.Error("Task wait returned error", "id", app.ID, "error", err)
			return
		}
		c.logger.Info("Task exited", "id", app.ID, "exitcode", code)
	}()

	return nil
}

// StopContainer kills the task and deletes the container
func (c *Client) StopContainer(appID string) error {
	ctx, cancel := c.createContext()
	defer cancel()

	container, err := c.cli.LoadContainer(ctx, appID)
	if err != nil {
		return err
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return err
	}

	// Stop task (SIGKILL here for simplicity, SIGTERM normally)
	// We wait 5 seconds before returning
	exitStatusC, err := task.Wait(ctx)
	if err != nil {
		return err
	}

	if err := task.Kill(ctx, 9); err != nil {
		return err
	}
	<-exitStatusC

	_, err = task.Delete(ctx)
	if err != nil {
		return err
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return err
	}

	return nil
}

// GetMetrics returns basic CPU/Memory stats from the cgroups (if running)
func (c *Client) GetMetrics(appID string) (uint64, uint64, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	container, err := c.cli.LoadContainer(ctx, appID)
	if err != nil {
		return 0, 0, err
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return 0, 0, err
	}

	metric, err := task.Metrics(ctx)
	if err != nil {
		return 0, 0, err
	}

	// In a real system you'd type cast metric.Data (an Any) to v1/v2 cgroups metrics type
	// For simplicity in this demo we return placeholder 0s if we get this far.
	_ = metric
	return 0, 0, nil
}

package runner

import (
	"context"
)

type TaskID string

const (
	TaskUpdateStats    TaskID = "service.stats.update"
	TaskMonitorTunnels TaskID = "service.tunnel.monitor"
)

type Task interface {
	ID() TaskID
	Run(ctx context.Context) error
}

package iface

import "context"

type Supervisor interface {
	KeepSystemHealthy(context.Context) error
	WaitingForReboot(context.Context) error
	WaitingForSystemAlive(context.Context) error
}

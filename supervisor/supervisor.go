package supervisor

import (
	"context"
	"fmt"
	"time"

	comfyui "github.com/sko00o/comfyui-go"
	"github.com/sko00o/comfyui-go/iface"
	"github.com/sko00o/comfyui-go/logger"
)

var _ iface.Supervisor = (*Supervisor)(nil)

type Supervisor struct {
	*comfyui.Client
	Logger logger.LoggerExtend

	// RAMFreeThreshold is the threshold of free RAM usage
	RAMFreeThreshold float64
	// VRAMFreeThreshold is the threshold of free VRAM usage
	VRAMFreeThreshold float64
	// TorchVRAMFreeThreshold is the threshold of free torch VRAM usage
	TorchVRAMFreeThreshold float64
}

type Option func(d *Supervisor)

func WithLogger(l logger.LoggerExtend) Option {
	return func(d *Supervisor) {
		d.Logger = l
	}
}

func WithRAMFreeThreshold(threshold float64) Option {
	return func(d *Supervisor) {
		d.RAMFreeThreshold = threshold
	}
}

func WithVRAMFreeThreshold(threshold float64) Option {
	return func(d *Supervisor) {
		d.VRAMFreeThreshold = threshold
	}
}

func WithTorchVRAMFreeThreshold(threshold float64) Option {
	return func(d *Supervisor) {
		d.TorchVRAMFreeThreshold = threshold
	}
}

func NewSupervisor(client *comfyui.Client, opts ...Option) *Supervisor {
	s := &Supervisor{
		Client: client,
		Logger: logger.NewStd(),

		RAMFreeThreshold:       0.1,
		VRAMFreeThreshold:      0.2,
		TorchVRAMFreeThreshold: 0.1,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (d *Supervisor) KeepSystemHealthy(ctx context.Context) error {
	if d.IsQueueEmpty() && d.IsSystemHealthy() {
		return nil
	}

	return d.WaitingForReboot(ctx)
}

func (d *Supervisor) WaitingForReboot(ctx context.Context) error {
	d.Logger.Infof("system start reboot...")
	_ = d.Reboot() // ignore any response
	return d.WaitingForSystemAlive(ctx)
}

func (d *Supervisor) WaitingForSystemAlive(ctx context.Context) error {
	for i := 0; ; i++ {
		if d.IsSystemAlive() {
			d.Logger.Infof("system is alive")
			return nil
		}
		d.Logger.Infof("waiting for system up... %d", i)
		select {
		case <-ctx.Done():
			d.Logger.Warnf("waiting for system up canceled")
			return fmt.Errorf("canceled")
		case <-time.After(time.Second * 3):
		}
	}
}

func (d *Supervisor) IsSystemAlive() bool {
	_, err := d.Stats()
	return err == nil
}

func (d *Supervisor) IsSystemHealthy() bool {
	resp, err := d.Stats()
	if err != nil {
		d.Logger.Warnf("system stats: %v", err)
		return false
	}

	if resp.System.RAMFree < int(float64(resp.System.RAMTotal)*d.RAMFreeThreshold) {
		d.Logger.Warnf("system ram is low: %d/%d", resp.System.RAMFree, resp.System.RAMTotal)
		return false
	}

	vramTotal, vramFree := 0, 0
	torchVRAMTotal, torchVRAMFree := 0, 0
	for _, device := range resp.Devices {
		vramTotal += device.VRAMTotal
		vramFree += device.VRAMFree
		torchVRAMTotal += device.TorchVRAMTotal
		torchVRAMFree += device.TorchVRAMFree
	}
	if vramFree < int(float64(vramTotal)*d.VRAMFreeThreshold) {
		d.Logger.Warnf("system vram is low: %d/%d", vramFree, vramTotal)
		return false
	}
	if torchVRAMFree < int(float64(torchVRAMTotal)*d.TorchVRAMFreeThreshold) {
		d.Logger.Warnf("system torch vram is low: %d/%d", torchVRAMFree, torchVRAMTotal)
		return false
	}
	d.Logger.Infof("system is healthy, free vram: %d/%d, free torch vram: %d/%d", vramFree, vramTotal, torchVRAMFree, torchVRAMTotal)
	return true
}

func (d *Supervisor) IsQueueEmpty() bool {
	resp, err := d.GetPrompt()
	if err != nil {
		d.Logger.Warnf("queue empty check: %v", err)
		return false
	}
	remain := resp.ExecInfo.QueueRemaining
	if remain != 0 {
		d.Logger.Warnf("queue empty check: not empty, remain: %d", remain)
		return false
	}
	return true
}

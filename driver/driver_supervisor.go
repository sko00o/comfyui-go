package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/sko00o/comfyui-go/iface"
)

var _ iface.Supervisor = (*Driver)(nil)

func (d *Driver) KeepSystemHealthy(ctx context.Context) error {
	if d.IsQueueEmpty() && d.IsSystemHealthy() {
		return nil
	}

	return d.WaitingForReboot(ctx)
}

func (d *Driver) WaitingForReboot(ctx context.Context) error {
	d.Logger.Infof("system start reboot...")
	_ = d.Reboot() // ignore any response
	return d.WaitingForSystemAlive(ctx)
}

func (d *Driver) WaitingForSystemAlive(ctx context.Context) error {
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

func (d *Driver) IsSystemAlive() bool {
	_, err := d.Stats()
	return err == nil
}

func (d *Driver) IsSystemHealthy() bool {
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

func (d *Driver) IsQueueEmpty() bool {
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

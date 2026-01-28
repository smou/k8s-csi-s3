package nodeserver_test

import (
	"context"
	"sync"

	"github.com/smou/k8s-csi-s3/pkg/driver/mount"
)

type FakeMountProvider struct {
	mu      sync.Mutex
	mounted map[string]bool

	lastMount mount.MountRequest
}

func NewFakeMountProvider() *FakeMountProvider {
	return &FakeMountProvider{
		mounted: make(map[string]bool),
	}
}

func (f *FakeMountProvider) IsMounted(targetPath string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mounted[targetPath], nil
}

func (f *FakeMountProvider) Mount(ctx context.Context, req mount.MountRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mounted[req.TargetPath] = true
	f.lastMount = req
	return nil
}

func (f *FakeMountProvider) Unmount(ctx context.Context, targetPath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.mounted, targetPath)
	return nil
}

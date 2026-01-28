package nodeserver_test

import (
	"context"
	"sync"

	"github.com/smou/k8s-csi-s3/pkg/driver/mount"
)

type FakeMountProvider struct {
	mu sync.Mutex

	mounted     map[string]bool
	mountErr    error
	unmountErr  error
	lastMount   *mount.MountRequest
	lastUnmount string
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
	f.lastMount = &req
	if f.mountErr != nil {
		return f.mountErr
	}
	f.mounted[req.TargetPath] = true
	return nil
}

func (f *FakeMountProvider) Unmount(ctx context.Context, targetPath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastUnmount = targetPath
	if f.unmountErr != nil {
		return f.unmountErr
	}
	delete(f.mounted, targetPath)
	return nil
}

package mount_test

import (
	"sync"

	"k8s.io/mount-utils"
)

type FakeMounter struct {
	mount.Interface
	mu      sync.Mutex
	mounted map[string]bool
}

func NewFakeMounter() *FakeMounter {
	return &FakeMounter{
		mounted: make(map[string]bool),
	}
}

func (f *FakeMounter) IsLikelyNotMountPoint(path string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return !f.mounted[path], nil
}

func (f *FakeMounter) Unmount(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.mounted, path)
	return nil
}

/* unbenutzte Methoden */
func (f *FakeMounter) Mount(_, _ string, _ string, _ []string) error {
	return nil
}
func (f *FakeMounter) List() ([]mount.MountPoint, error) {
	return nil, nil
}
func (f *FakeMounter) IsMountPoint(string) (bool, error) {
	return false, nil
}

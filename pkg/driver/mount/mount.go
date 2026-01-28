package mount

import "context"

type Provider interface {
	// Mount mountet das Volume auf den Zielpfad
	Mount(ctx context.Context, req MountRequest) error

	// Unmount entfernt das Mount
	Unmount(ctx context.Context, targetPath string) error

	// IsMounted prüft Idempotenz
	IsMounted(targetPath string) (bool, error)
}

type MountRequest struct {
	TargetPath string

	Bucket   string
	Endpoint string
	Region   string

	AccessKey string
	SecretKey string

	ReadOnly bool

	Options map[string]string
}

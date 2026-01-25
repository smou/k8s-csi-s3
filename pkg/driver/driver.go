/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"fmt"
	"net"
	"os"

	"github.com/smou/k8s-csi-s3/pkg/driver/version"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	driverName    = "de.smou.s3.csi"
	driverVersion = "v1.35.0"

	unixSocketPerm                  = os.FileMode(0700) // only owner can write and read.
	grpcServerMaxReceiveMessageSize = 1024 * 1024 * 2   // 2MB
)

type Driver struct {
	Endpoint string
	Srv      *gprc.Server
	NodeId   string

	ClientSet kubernetes.Interface
	csi.UnimplementedIdentityServer
	csi.UnimplementedControllerServer
}

type driver struct {
	driver   *csicommon.CSIDriver
	endpoint string

	ids *csicommon.DefaultIdentityServer
	ns  *csicommon.DefaultNodeServer
	cs  *csicommon.DefaultControllerServer
}

func newDriver(endpoint string, nodeID string) (*Driver, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create kubernetes clientset: %w", err)
	}
	kubernetesVersion, err := kubernetesVersion(clientset)
	if err != nil {
		klog.Errorf("failed to get kubernetes version: %v", err)
	}
	version := version.GetVersion()
	klog.Infof("Driver version: %v, Git commit: %v, build date: %v, nodeID: %v, mount-s3 version: %v, kubernetes version: %v",
		version.DriverVersion, version.GitCommit, version.BuildDate, nodeID, driverVersion, kubernetesVersion)

	// TODO mounter erstellen
	// mpMounter := mpmounter.New()

	return &Driver{
		Endpoint:  endpoint,
		NodeId:    nodeID,
		ClientSet: clientset,
	}, nil
}

func kubernetesVersion(clientset *kubernetes.Clientset) (string, error) {
	version, err := clientset.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("cannot get kubernetes server version: %w", err)
	}

	return version.String(), nil
}

// // New initializes the driver
// func New(nodeID string, endpoint string) (*driver, error) {
// 	d := csicommon.NewCSIDriver(driverName, vendorVersion, nodeID)
// 	if d == nil {
// 		glog.Fatalln("Failed to initialize CSI Driver.")
// 	}

// 	s3Driver := &driver{
// 		endpoint: endpoint,
// 		driver:   d,
// 	}
// 	return s3Driver, nil
// }

func newIdentityServer(d *csicommon.CSIDriver) *csicommon.DefaultIdentityServer {
	return csicommon.NewDefaultIdentityServer(d)
}

func (s3 *driver) newControllerServer(d *csicommon.CSIDriver) *csicommon.DefaultControllerServer {
	return csicommon.NewDefaultControllerServer(d)
}

func (s3 *driver) newNodeServer(d *csicommon.CSIDriver) *csicommon.DefaultNodeServer {
	return csicommon.NewDefaultNodeServer(d)
}

func (d *Driver) Run() error {
	scheme, addr, err := ParseEndpoint(d.Endpoint)
	if err != nil {
		return err
	}

	listener, err := net.Listen(scheme, addr)
	if err != nil {
		return err
	}
	if scheme == "unix" {
		// Go's `net` package does not support specifying permissions on Unix sockets it creates.
		// There are two ways to change permissions:
		// 	 - Using `syscall.Umask` before `net.Listen`
		//   - Calling `os.Chmod` after `net.Listen`
		// The first one is not nice because it affects all files created in the process,
		// the second one has a time-window where the permissions of Unix socket would depend on `umask`
		// between `net.Listen` and `os.Chmod`. Since we don't start accepting connections on the socket until
		// `grpc.Serve` call, we should be fine with `os.Chmod` option.
		// See https://github.com/golang/go/issues/11822#issuecomment-123850227.
		if err := os.Chmod(addr, unixSocketPerm); err != nil {
			klog.Errorf("Failed to change permissions on unix socket %s: %v", addr, err)
			return fmt.Errorf("Failed to change permissions on unix socket %s: %v", addr, err)
		}
	}

	logErr := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			klog.Errorf("GRPC error: %v", err)
		}
		return resp, err
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logErr),
		grpc.MaxRecvMsgSize(grpcServerMaxReceiveMessageSize),
	}
	d.Srv = grpc.NewServer(opts...)

	csi.RegisterIdentityServer(d.Srv, d)
	csi.RegisterControllerServer(&d.Srv, d)
	csi.RegisterNodeServer(&d.Srv, d)

	klog.Infof("Listening for connections on address: %#v", listener.Addr())

	return d.Srv.Serve(listener)

	// s3.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	// s3.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})

	// // Create GRPC servers
	// ids := csicommon.NewDefaultIdentityServer(s3.driver)
	// s3.ns = s3.newNodeServer(s3.driver)
	// s3.cs = s3.newControllerServer(s3.driver)

	// s := csicommon.NewNonBlockingGRPCServer()
	// s.Start(s3.endpoint, csi.NewDefaultIdentityServer(d), s3.cs, s3.ns)
	// s.Wait()
}

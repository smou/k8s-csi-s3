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

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/smou/k8s-csi-s3/pkg/config"
	"github.com/smou/k8s-csi-s3/pkg/driver"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	endpoint    = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID      = flag.String("nodeid", "", "node id")
	mountBinary = flag.String("mountBinary", "/usr/bin/mountpoint-s3", "s3 mount binary path")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	config, err := config.InitConfig(ctx)
	if err != nil {
		log.Fatalf("Error loading DriverConfig: %v", err)
	}
	config.Endpoint = *endpoint
	config.NodeID = *nodeID
	config.MountBinary = *mountBinary

	driver, err := driver.NewDriver(config)
	if err != nil {
		log.Fatalf("Error run Driver: %v", err)
	}
	go func() {
		if err := driver.Run(); err != nil {
			log.Fatalf("driver error: %v", err)
		}
	}()
	<-ctx.Done()
	driver.Stop()
	os.Exit(0)
}

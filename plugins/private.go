// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package plugins

import "net/rpc"

type builderClient struct {
	client *rpc.Client
}

type builderServer struct {
	Impl Builder
}

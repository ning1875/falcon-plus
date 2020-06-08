// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package g

import (
	"errors"
	"log"
	"math"
	"net/rpc"
	"sync"
	"time"

	"github.com/toolkits/net"
)

// 简单封装rpc.Cilent
type SingleConnRpcClient struct {
	sync.Mutex
	rpcClient *rpc.Client
	RpcServer string
	Timeout   time.Duration
}

// 关闭rpc
func (this *SingleConnRpcClient) close() {
	if this.rpcClient != nil {
		this.rpcClient.Close()
		this.rpcClient = nil
	}
}

// 保证rpc存在,为空则重新创建
func (this *SingleConnRpcClient) serverConn() error {
	if this.rpcClient != nil {
		return nil
	}

	var err error
	var retry int = 1

	for {
		if this.rpcClient != nil {
			return nil
		}
		// 根据timeout和server地址去连接rpc的server
		this.rpcClient, err = net.JsonRpcClient("tcp", this.RpcServer, this.Timeout)
		if err != nil {
			log.Printf("dial %s fail: %v", this.RpcServer, err)
			if retry > 3 {
				return err
			}
			time.Sleep(time.Duration(math.Pow(2.0, float64(retry))) * time.Second)
			retry++
			continue
		}
		return err
	}
}

// rpc client调用hbs函数
func (this *SingleConnRpcClient) Call(method string, args interface{}, reply interface{}) error {
	// 加锁保证一个agent只与server有一个连接,保证性能
	this.Lock()
	defer this.Unlock()
	// 保证rpc连接可用
	err := this.serverConn()
	if err != nil {
		return err
	}

	timeout := time.Duration(10 * time.Second)
	done := make(chan error, 1)

	go func() {
		err := this.rpcClient.Call(method, args, reply)
		done <- err
	}()
	// 超时控制
	select {
	case <-time.After(timeout):
		log.Printf("[WARN] rpc call timeout %v => %v", this.rpcClient, this.RpcServer)
		this.close()
		return errors.New(this.RpcServer + " rpc call timeout")
	case err := <-done:
		if err != nil {
			this.close()
			return err
		}
	}

	return nil
}

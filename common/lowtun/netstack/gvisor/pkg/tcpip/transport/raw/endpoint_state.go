// Copyright 2018 The gVisor Authors.
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

package raw

import (
	"context"
	"time"

	"github.com/yaklang/yaklang/common/lowtun/netstack/gvisor/pkg/tcpip"
	"github.com/yaklang/yaklang/common/lowtun/netstack/gvisor/pkg/tcpip/stack"
)

// saveReceivedAt is invoked by stateify.
func (p *rawPacket) saveReceivedAt() int64 {
	return p.receivedAt.UnixNano()
}

// loadReceivedAt is invoked by stateify.
func (p *rawPacket) loadReceivedAt(_ context.Context, nsec int64) {
	p.receivedAt = time.Unix(0, nsec)
}

// afterLoad is invoked by stateify.
func (e *endpoint) afterLoad(ctx context.Context) {
	if e.stack.IsSaveRestoreEnabled() {
		e.stack.RegisterRestoredEndpoint(e)
	} else {
		stack.RestoreStackFromContext(ctx).RegisterRestoredEndpoint(e)
	}
}

// beforeSave is invoked by stateify.
func (e *endpoint) beforeSave() {
	e.setReceiveDisabled(true)
	e.stack.RegisterResumableEndpoint(e)
}

// Restore implements tcpip.RestoredEndpoint.Restore.
func (e *endpoint) Restore(s *stack.Stack) {
	e.setReceiveDisabled(false)
	e.net.Resume(s)
	if e.stack.IsSaveRestoreEnabled() {
		e.ops.InitHandler(e, e.stack, tcpip.GetStackSendBufferLimits, tcpip.GetStackReceiveBufferLimits)
		return
	}

	e.stack = s
	e.ops.InitHandler(e, e.stack, tcpip.GetStackSendBufferLimits, tcpip.GetStackReceiveBufferLimits)

	if e.associated {
		netProto := e.net.NetProto()
		if err := e.stack.RegisterRawTransportEndpoint(netProto, e.transProto, e); err != nil {
			panic("RegisterRawTransportEndpoint failed during restore")
		}
	}
}

// Resume implements tcpip.ResumableEndpoint.Resume.
func (e *endpoint) Resume() {
	e.setReceiveDisabled(false)
}

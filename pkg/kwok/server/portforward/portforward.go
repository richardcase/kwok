/*
Copyright 2016 The Kubernetes Authors.

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

package portforward

import (
	"context"
	"io"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/util/wsstream"

	"sigs.k8s.io/kwok/pkg/log"
)

// PortForwarder knows how to forward content from a data stream to/from a port
// in a pod.
type PortForwarder interface {
	// PortForwarder copies data between a data stream and a port in a pod.
	PortForward(ctx context.Context, podName, podNamespace string, uid types.UID, port int32, stream io.ReadWriteCloser) error
}

// ServePortForward handles a port forwarding request.  A single request is
// kept alive as long as the client is still alive and the connection has not
// been timed out due to idleness. This function handles multiple forwarded
// connections; i.e., multiple `curl http://localhost:8888/` requests will be
// handled by a single invocation of ServePortForward.
func ServePortForward(ctx context.Context, w http.ResponseWriter, req *http.Request, portForwarder PortForwarder, podName, podNamespace string, uid types.UID, portForwardOptions *V4Options, idleTimeout time.Duration, streamCreationTimeout time.Duration, supportedProtocols []string) {
	var err error
	if wsstream.IsWebSocketRequest(req) {
		err = handleWebSocketStreams(ctx, req, w, portForwarder, podName, podNamespace, uid, portForwardOptions, idleTimeout)
	} else {
		err = handleHTTPStreams(ctx, req, w, portForwarder, podName, podNamespace, uid, supportedProtocols, idleTimeout, streamCreationTimeout)
	}

	if err != nil {
		logger := log.FromContext(req.Context())
		logger.Error("handle stream", err)
		return
	}
}

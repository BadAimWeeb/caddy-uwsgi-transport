// Copyright 2023 Xinhe Wang
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

package uwsgi

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
)

func init() {
	caddy.RegisterModule(Transport{})
}

type Transport struct {
	UWSGIParams map[string]string `json:"uwsgi_params,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (Transport) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.reverse_proxy.transport.uwsgi",
		New: func() caddy.Module { return new(Transport) },
	}
}

var headerNameReplacer = strings.NewReplacer("-", "_")

func writeBlockVar(buffer *bytes.Buffer, s string) {
	b := []byte(s)
	binary.Write(buffer, binary.LittleEndian, uint16(len(b)))
	buffer.Write(b)
}

// generateBlockVars returns the packet body of WSGI block vars generated from http.Request.
func generateBlockVars(req *http.Request, t Transport) (*bytes.Buffer, error) {
	serverName, serverPort, err := net.SplitHostPort(req.Host)
	if err != nil {
		serverName = req.Host
	}
	if serverPort == "" {
		if req.TLS == nil {
			serverPort = "80"
		} else {
			serverPort = "443"
		}
	}

	httpsConn := ""
	if req.TLS != nil {
		httpsConn = "on"
	}

	vars := map[string]string{
		"QUERY_STRING":    req.URL.RawQuery,
		"REQUEST_METHOD":  req.Method,
		"CONTENT_TYPE":    req.Header.Get("Content-Type"),
		"CONTENT_LENGTH":  req.Header.Get("Content-Length"),

		"REQUEST_URI":     req.RequestURI,
		"PATH_INFO":       req.URL.Path,
		"SERVER_PROTOCOL": req.Proto,
		"REQUEST_SCHEME":  req.URL.Scheme,
		"HTTPS":		   httpsConn,
		
		"REMOTE_ADDR":     strings.Split(req.RemoteAddr, ":")[0],
		"REMOTE_PORT":     strings.Split(req.RemoteAddr, ":")[1],
		"SERVER_PORT":     serverPort,
		"SERVER_NAME":     serverName,

		"HTTP_HOST":       req.Host,
	}
	if req.TLS != nil {
		vars["HTTPS"] = "on"
	}
	for name, value := range req.Header {
		vars["HTTP_"+headerNameReplacer.Replace(strings.ToUpper(name))] = strings.Join(value, ", ")
	}

	for name, value := range t.UWSGIParams {
		vars[name] = value
	}

	var packetBody bytes.Buffer
	for key, val := range vars {
		writeBlockVar(&packetBody, key)
		writeBlockVar(&packetBody, val)
	}
	return &packetBody, nil
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	network, address := "tcp", req.URL.Host
	if dialInfo, ok := reverseproxy.GetDialInfo(ctx); ok {
		network = dialInfo.Network
		address = dialInfo.Address
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	blockVars, err := generateBlockVars(req, t)
	if err != nil {
		return nil, err
	}

	conn.Write([]byte{0})                                            // modifier1
	binary.Write(conn, binary.LittleEndian, uint16(blockVars.Len())) // datasize
	conn.Write([]byte{0})                                            // modifier2
	io.Copy(conn, blockVars)                                         // packet body

	if req.Body != nil {
		io.Copy(conn, req.Body)
		req.Body.Close()
	}

	return http.ReadResponse(bufio.NewReader(conn), req)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (t *Transport) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.NextArg() {
			return nil
		}

		if d.Val() == "uwsgi_param" {
			if !d.NextArg() {
				return d.ArgErr()
			}

			key := d.Val()

			if !d.NextArg() {
				return d.ArgErr()
			}

			value := d.Val()

			t.UWSGIParams[key] = value
		} else {
			return d.Errf("unknown subdirective %s", d.Val())
		}
	}

	return nil
}

func (t *Transport) 

var (
	_ http.RoundTripper     = (*Transport)(nil)
	_ caddyfile.Unmarshaler = (*Transport)(nil)
)

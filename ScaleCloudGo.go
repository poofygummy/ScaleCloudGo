package ScaleCloudGo

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"tailscale.com/tsnet"
	"tailscale.com/types/logger"
	// "tailscale.com/net/netmon"
	// "github.com/wlynxg/anet"
	// "strconv"
	// "io"
	// "bytes"
)

// TailScale Netstack usage and OAuth preset
const tsClientID = "kRkUEfX6op11CNTRL"
const tsClientSecret = "tskey-client-kRkUEfX6op11CNTRL-RYXbEQR7XmfneD1PrCKPnfc5DqKW6HT8E"

func init() {
	os.Setenv("TS_AUTHKEY", "")
	os.Setenv("TS_OAUTH_CLIENT_ID", tsClientID)
	os.Setenv("TS_OAUTH_CLIENT_SECRET", tsClientSecret)
}

// TailScale Address Helper
var _, tailscaleRange, _ = net.ParseCIDR("100.64.0.0/10")

func isTailscaleAddress(host string) bool {
	// 1. Check for the MagicDNS suffix
	if strings.HasSuffix(host, ".ts.net") {
		return true
	}
	// 2. Check for an IP
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	// 3. Check for the TailScale range
	if tailscaleRange.Contains(ip) {
		return true
	}
	return false
}

// TailScale Node Activation helper
var tsNode *tsnet.Server
var nodeMX sync.Mutex // The Mutual eXclusivity lock
func ensureTSNodeActive(hostname, stateDir string) error {
	nodeMX.Lock()
	defer nodeMX.Unlock()
	if tsNode == nil {
		tsNode = &tsnet.Server{
			Hostname:   hostname,
			Ephemeral:  false,
			Logf:       logger.Discard,
			Dir:        stateDir,
			ControlURL: "https://controlplane.tailscale.com",
		}
		os.Setenv("TS_LOGS_DIR", stateDir)
		err := tsNode.Start()
		if err != nil {
			tsNode = nil // Reset so we can try again later
			return err
		}
	}
	return nil
}

// --- THE ENGINE ---

var proxyServer *http.Server
var assignedAddr string
var assignedPort int
var proxyMX sync.Mutex // The Mutual eXclusivity lock
var baseDialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second}

func StartProxy(hostname, stateDir string) (int, error) {
	// 0. Function lock
	proxyMX.Lock()
	defer proxyMX.Unlock()

	// 1. If proxy already running, return existing port
	if proxyServer != nil {
		return assignedPort, nil
	}

	// 2. Start the listener and get its data
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	assignedAddr = listener.Addr().String()
	assignedPort = listener.Addr().(*net.TCPAddr).Port

	// 3. Create the proxy server
	// a) goproxy logic init
	proxy := goproxy.NewProxyHttpServer()
	// b) smart dialer: routes Tailscale addresses through tsnet, everything else normally
	smartDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, _ := net.SplitHostPort(addr)
		if isTailscaleAddress(host) {
			err := ensureTSNodeActive(hostname, stateDir)
			if err != nil {
				return nil, err
			}
			state, err := tsNode.Up(ctx)
			if err != nil {
				return nil, err
			}
			_ = state
			return tsNode.Dial(ctx, network, addr) // tsnet
		}
		return baseDialer.DialContext(ctx, network, addr) // normal
	}

	// c) HTTP requests use the smart dialer via Transport
	proxy.Tr = &http.Transport{
		DialContext: smartDial,
	}
	// d) CONNECT tunnel handler — required for HTTPS
	//    iOS sends CONNECT host:443 for all https:// requests going through a proxy.
	//    Without this goproxy rejects CONNECT and iOS gets kCFErrorHTTPSProxyConnectionFailure (error 310).
	proxy.ConnectDial = func(network, addr string) (net.Conn, error) {
		return smartDial(context.Background(), network, addr)
	}
	// d) server struct construction
	proxyServer = &http.Server{
		Handler: proxy,
		Addr:    assignedAddr, // manually telling it its own address
	}
	// e) launching server struct via go func so we can move past
	go func() {
		proxyServer.Serve(listener)
	}()

	// 4. Return the port
	return assignedPort, nil
}

func StopProxy() error {
	proxyMX.Lock()
	nodeMX.Lock()
	defer proxyMX.Unlock()
	defer nodeMX.Unlock()
	var err error
	if tsNode != nil {
		tsErr := tsNode.Close()
		tsNode = nil
		if tsErr != nil {
			err = tsErr
		}
	} // TS errors
	if proxyServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pErr := proxyServer.Shutdown(ctx)
		proxyServer = nil
		assignedPort = 0
		if pErr != nil {
			err = pErr
		}
	} // proxy errors
	return err
}

package ScaleCloudGo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"tailscale.com/tsnet"
	// "tailscale.com/types/logger"
	// "tailscale.com/net/netmon"
	// "github.com/wlynxg/anet"
	// "strconv"
	// "io"
	// "bytes"
)

// TailScale Netstack usage and OAuth preset
const tsClientID = "ku5mMZ9zKW11CNTRL"

// ?ephemeral=false: OAuth-generated auth keys are ephemeral by default.
// The Ephemeral field on tsnet.Server does NOT control this —
// the OAuth resolver (feature/oauthkey) parses these query params from the
// secret string itself and passes them to the Tailscale CreateKey API.
// We do NOT request preauthorized=true here: that capability requires the
// OAuth key to have been granted the preauthorized scope in the Tailscale
// admin console. If that scope is absent the CreateKey call would fail or
// fall back to creating an ephemeral node. Omitting it lets the key be
// created as a normal (non-ephemeral, non-preauthorized) node.
const tsClientSecret = "tskey-client-ku5mMZ9zKW11CNTRL-Q3C62fWbKEKTo7aDTC6XDKaU2jpPHhCEf?ephemeral=false"

// --- DIAGNOSTICS ---

var logMX sync.Mutex
var logLines []string
var lastNodeState string

func tsLog(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	// Print immediately to stdout so idevicedebug captures it live.
	fmt.Println("[tsnet]", line)
	logMX.Lock()
	logLines = append(logLines, line)
	if len(logLines) > 200 {
		logLines = logLines[len(logLines)-200:]
	}
	logMX.Unlock()
}

// GetLogs returns the last known node state token on the first line ("STATE:<token>"),
// followed by all captured tsnet log lines. Clears the log buffer after read.
// Swift splits on the first newline to get state and logs separately.
func GetLogs() string {
	logMX.Lock()
	defer logMX.Unlock()
	state := lastNodeState
	if state == "" {
		state = "Unknown"
	}
	result := "STATE:" + state + "\n" + strings.Join(logLines, "\n")
	logLines = nil // clear after read
	return result
}

func init() {
	os.Setenv("TS_AUTHKEY", "")
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

// nodeStatusGate checks the current Tailscale node state once.
// Returns "NodeReady" on success, or "<State>\n<logs>" for all error states.
// Swift splits on the first newline: left = state string, right = log dump.
func nodeStatusGate() string {
	logs := func() string {
		logMX.Lock()
		defer logMX.Unlock()
		return strings.Join(logLines, "\n")
	}

	nodeMX.Lock()
	srv := tsNode
	nodeMX.Unlock()
	if srv == nil {
		logMX.Lock()
		lastNodeState = "NodeStarting"
		logMX.Unlock()
		return "NodeStarting\n" + logs()
	}

	lc, lcErr := srv.LocalClient()
	if lcErr != nil {
		logMX.Lock()
		lastNodeState = "OtherError"
		logMX.Unlock()
		return "OtherError\n" + logs()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	st, stErr := lc.Status(ctx)

	var selfTags []string
	if st != nil && st.Self != nil && st.Self.Tags != nil {
		selfTags = st.Self.Tags.AsSlice()
	}
	if st != nil {
		tsLog("[gate] BackendState=%s TailscaleIPs=%v Tags=%v",
			st.BackendState, st.TailscaleIPs, selfTags)
	}

	var state string
	switch {
	case stErr != nil:
		state = "OtherError"
	case st.BackendState == "Running" && contains(selfTags, "tag:scalecloud-ios"):
		state = "NodeReady"
	case st.BackendState == "Running" && contains(selfTags, "tag:scalecloud-ios-pending"):
		state = "NeedsRetag"
	case st.BackendState == "NeedsMachineAuth":
		state = "NeedsAuth"
	case st.BackendState == "Starting" || st.BackendState == "Connecting" || st.BackendState == "":
		state = "NodeStarting"
	default:
		state = "OtherError"
	}
	logMX.Lock()
	lastNodeState = state
	logMX.Unlock()
	if state == "NodeReady" {
		return "NodeReady"
	}
	return state + "\n" + logs()
}

func contains(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
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
		// Set TS_LOGS_DIR before Start() — logpolicy.LogsDir() reads it at the
		// top of Start(). If it is empty when Start() runs, logpolicy falls back
		// to os.UserCacheDir() which requires $HOME. Under idevicedebug (and any
		// non-SpringBoard launcher) $HOME is unset, so that fallback fails and
		// logpolicy panics with "no safe place found to store log state".
		// Also pin HOME so that any other Go stdlib path that reads it has a
		// writable sandbox location rather than whatever the launcher left.
		os.Setenv("TS_LOGS_DIR", stateDir)
		if os.Getenv("HOME") == "" {
			os.Setenv("HOME", stateDir)
		}
		tsNode = &tsnet.Server{
			Hostname: hostname,
			// Ephemeral: false here does nothing when ClientSecret is used —
			// the ephemeral flag is controlled by the ?ephemeral=false query
			// param encoded in the tsClientSecret string above.
			Logf:          tsLog,
			Dir:           stateDir,
			ControlURL:    "https://controlplane.tailscale.com",
			ClientSecret:  tsClientSecret,
			AdvertiseTags: []string{"tag:scalecloud-ios-pending"},
		}
		err := tsNode.Start()
		if err != nil {
			tsNode = nil // Reset so we can try again later
			return err
		}
		// Start a background goroutine that polls the IPN state machine and
		// prints it to stdout every 5 s so we can see it live in idevicedebug.
		go func(srv *tsnet.Server) {
			for {
				time.Sleep(5 * time.Second)
				nodeMX.Lock()
				alive := tsNode == srv
				nodeMX.Unlock()
				if !alive {
					return
				}
				lc, err := srv.LocalClient()
				if err != nil {
					fmt.Printf("[tsnet-poll] LocalClient error: %v\n", err)
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				st, err := lc.Status(ctx)
				cancel()
				if err != nil {
					fmt.Printf("[tsnet-poll] Status error: %v\n", err)
					continue
				}
				fmt.Printf("[tsnet-poll] BackendState=%s AuthURL=%q IPs=%v\n",
					st.BackendState, st.AuthURL, st.TailscaleIPs)
			}
		}(tsNode)
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
				tsLog("ensureTSNodeActive: Start() failed: %v", err)
				logMX.Lock()
				lastNodeState = "NodeStarting"
				logMX.Unlock()
				return nil, fmt.Errorf("NodeStarting")
			}
			// Gate: return state to Swift if not ready; Swift retries every 10s.
			if st := nodeStatusGate(); st != "NodeReady" {
				return nil, fmt.Errorf("%s", st)
			}
			// Use context.Background() for Up and Dial so no upstream timeout
			// can cancel the connection and leave the node in a broken state.
			nodeMX.Lock()
			srv := tsNode
			nodeMX.Unlock()
			_, err = srv.Up(context.Background())
			if err != nil {
				return nil, fmt.Errorf("OtherError\n%s", GetLogs())
			}
			return srv.Dial(context.Background(), network, addr) // tsnet
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

package localbinary

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	// Timeout where we will bail if we're not able to properly contact the
	// plugin server.
	defaultTimeout = 10 * time.Second
	CoreDrivers    = []string{"k8s"}
)

const (
	pluginOut           = "(%s) %s"
	pluginErr           = "(%s) DBG | %s"
	PluginEnvKey        = "MACHINE_PLUGIN_TOKEN"
	PluginEnvVal        = "42"
	PluginEnvDriverName = "MACHINE_PLUGIN_DRIVER_NAME"
	PluginBuildIn       = "MACHINE_PLUGIN_BUILD_IN"
)

type PluginStreamer interface {
	// Return a channel for receiving the output of the stream line by
	// line.
	//
	// It happens to be the case that we do this all inside of the main
	// plugin struct today, but that may not be the case forever.
	AttachStream(*bufio.Scanner) <-chan string
}

type PluginServer interface {
	// Get the address where the plugin server is listening.
	Address() (string, error)

	// Serve kicks off the plugin server.
	Serve() error

	// Close shuts down the initialized server.
	Close() error
}

type McnBinaryExecutor interface {
	// Execute the driver plugin.  Returns scanners for plugin binary
	// stdout and stderr.
	Start() (*bufio.Scanner, *bufio.Scanner, error)

	// Stop reading from the plugins in question.
	Close() error
}

// DriverPlugin interface wraps the underlying mechanics of starting a driver
// plugin server and then figuring out where it can be dialed.
type DriverPlugin interface {
	PluginServer
	PluginStreamer
}

type Plugin struct {
	Ctx         context.Context
	Executor    McnBinaryExecutor
	Addr        string
	MachineName string
	addrCh      chan string
	stopCh      chan struct{}
	timeout     time.Duration
}

type Executor struct {
	ctx                        context.Context
	pluginStdout, pluginStderr io.ReadCloser
	DriverName                 string
	DriverVersion              string
	cmd                        *exec.Cmd
	binaryPath                 string
}

type ErrPluginBinaryNotFound struct {
	driverName string
	driverPath string
}

func (e ErrPluginBinaryNotFound) Error() string {
	return fmt.Sprintf("Driver %q not found. Do you have the plugin binary %q accessible in your PATH?", e.driverName, e.driverPath)
}

// driverPath finds the path of a driver binary by its name.
func driverPath(driverName string) string {
	for _, coreDriver := range CoreDrivers {
		if coreDriver == driverName {
			return os.Args[0]
		}
	}
	return fmt.Sprintf("plugins/%s/%s/plugin", driverName, "v1_0_0")
}

// NewPlugin new plugin by driver name
func NewPlugin(ctx context.Context, driverName string) (*Plugin, error) {
	dp := driverPath(driverName)
	dpabs, _ := filepath.Abs(dp)
	_, err := os.Stat(dpabs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrPluginBinaryNotFound{driverName, dpabs}
		}
		return nil, errors.Wrap(err, "state plugin")
	}

	log.Debugf(ctx, "Found binary path at %s", dpabs)
	_ = os.Chmod(dpabs, os.ModePerm)

	return &Plugin{
		Ctx:    ctx,
		stopCh: make(chan struct{}),
		addrCh: make(chan string, 1),
		Executor: &Executor{
			ctx:           ctx,
			DriverVersion: "v1_0_0",
			DriverName:    driverName,
			binaryPath:    dpabs,
		},
	}, nil
}

func (lbe *Executor) Start() (*bufio.Scanner, *bufio.Scanner, error) {
	var err error

	log.Debugf(lbe.ctx, "Launching plugin server for driver %s", lbe.DriverName)

	lbe.cmd = exec.Command(lbe.binaryPath)

	lbe.pluginStdout, err = lbe.cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting cmd stdout pipe: %s", err)
	}

	lbe.pluginStderr, err = lbe.cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting cmd stderr pipe: %s", err)
	}

	outScanner := bufio.NewScanner(lbe.pluginStdout)
	errScanner := bufio.NewScanner(lbe.pluginStderr)

	os.Setenv(PluginEnvKey, PluginEnvVal)
	os.Setenv(PluginEnvDriverName, lbe.DriverName)
	if os.Args[0] == lbe.binaryPath {
		os.Setenv(PluginBuildIn, "true")
	}

	if err := lbe.cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("Error starting plugin binary: %s", err)
	}

	return outScanner, errScanner, nil
}

func (lbe *Executor) Close() error {
	if err := lbe.cmd.Wait(); err != nil {
		return fmt.Errorf("Error waiting for binary close: %s", err)
	}

	return nil
}

func stream(scanner *bufio.Scanner, streamOutCh chan<- string, stopCh <-chan struct{}) {
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			log.Warnf(context.Background(), "Scanning stream: %s", err)
		}
		select {
		case streamOutCh <- strings.Trim(line, "\n"):
		case <-stopCh:
			return
		}
	}
}

func (lbp *Plugin) AttachStream(scanner *bufio.Scanner) <-chan string {
	streamOutCh := make(chan string)
	go stream(scanner, streamOutCh, lbp.stopCh)
	return streamOutCh
}

func (lbp *Plugin) execServer() error {
	outScanner, errScanner, err := lbp.Executor.Start()
	if err != nil {
		return err
	}

	// Scan just one line to get the address, then send it to the relevant
	// channel.
	outScanner.Scan()
	addr := outScanner.Text()
	if err := outScanner.Err(); err != nil {
		return fmt.Errorf("Reading plugin address failed: %s", err)
	}

	lbp.addrCh <- strings.TrimSpace(addr)

	stdOutCh := lbp.AttachStream(outScanner)
	stdErrCh := lbp.AttachStream(errScanner)

	for {
		select {
		case out := <-stdOutCh:
			log.Infof(lbp.Ctx, pluginOut, lbp.MachineName, out)
		case err := <-stdErrCh:
			log.Debugf(lbp.Ctx, pluginErr, lbp.MachineName, err)
		case <-lbp.stopCh:
			if err := lbp.Executor.Close(); err != nil {
				return fmt.Errorf("Error closing local plugin binary: %s", err)
			}
			return nil
		}
	}
}

func (lbp *Plugin) Serve() error {
	return lbp.execServer()
}

func (lbp *Plugin) Address() (string, error) {
	if lbp.Addr == "" {
		if lbp.timeout == 0 {
			lbp.timeout = defaultTimeout
		}

		select {
		case lbp.Addr = <-lbp.addrCh:
			log.Debugf(lbp.Ctx, "Plugin server listening at address %s", lbp.Addr)
			close(lbp.addrCh)
			return lbp.Addr, nil
		case <-time.After(lbp.timeout):
			return "", fmt.Errorf("Failed to dial the plugin server in %s", lbp.timeout)
		}
	}
	return lbp.Addr, nil
}

func (lbp *Plugin) Close() error {
	close(lbp.stopCh)
	return nil
}

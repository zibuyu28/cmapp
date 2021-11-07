package localbinary

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var (
	// Timeout where we will bail if we're not able to properly contact the
	// plugin server.
	defaultTimeout = 10 * time.Second
	CoreDrivers    = []string{"k8s", "virtualbox"}
)

const (
	pluginOut              = "(%s) %s"
	pluginErr              = "(%s) DBG | %s"
	PluginEnvKey           = "MACHINE_PLUGIN_TOKEN"
	PluginEnvVal           = "42"
	PluginEnvDriverName    = "MACHINE_PLUGIN_DRIVER_NAME"
	PluginEnvDriverVersion = "MACHINE_PLUGIN_DRIVER_VERSION"
	PluginEnvDriverID      = "MACHINE_PLUGIN_DRIVER_ID"
	PluginBuildIn          = "MACHINE_PLUGIN_BUILD_IN"
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
	DriverID                   int
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
func driverPath(driverName, driverVersion string) string {
	for _, coreDriver := range CoreDrivers {
		if coreDriver == driverName {
			return os.Args[0]
		}
	}
	return fmt.Sprintf("plugins/%s/%s/plugin", driverName, driverVersion)
}

// NewPlugin new plugin by driver name
func NewPlugin(ctx context.Context, driverID int, driverName, driverVersion string) (*Plugin, error) {
	dp := driverPath(driverName, driverVersion)
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
			DriverVersion: driverVersion,
			DriverName:    driverName,
			DriverID:      driverID,
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
	for _, v := range os.Environ() {
		lbe.cmd.Env = append(lbe.cmd.Env, v)
	}
	outScanner := bufio.NewScanner(lbe.pluginStdout)
	errScanner := bufio.NewScanner(lbe.pluginStderr)

	lbe.cmd.Env = append(lbe.cmd.Env, fmt.Sprintf("%s=%s", PluginEnvKey, PluginEnvVal))
	lbe.cmd.Env = append(lbe.cmd.Env, fmt.Sprintf("%s=%s", PluginEnvDriverName, lbe.DriverName))
	lbe.cmd.Env = append(lbe.cmd.Env, fmt.Sprintf("%s=%s", PluginEnvDriverVersion, lbe.DriverVersion))
	lbe.cmd.Env = append(lbe.cmd.Env, fmt.Sprintf("%s=%d", PluginEnvDriverID, lbe.DriverID))
	abs, _ := filepath.Abs(os.Args[0])
	if abs == lbe.binaryPath {
		lbe.cmd.Env = append(lbe.cmd.Env, fmt.Sprintf("%s=%s", PluginBuildIn, "true"))
	}
	err = lbe.cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("Error starting plugin binary: %v", err)
	}

	return outScanner, errScanner, nil
}

func (lbe *Executor) Close() error {
	go func() {
		<-lbe.ctx.Done()
		_ = syscall.Kill(lbe.cmd.Process.Pid, syscall.SIGKILL)
	}()
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
	log.Debugf(lbp.Ctx, "start to get addr")
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
		case <-lbp.Ctx.Done():
			if err := lbp.Executor.Close(); err != nil {
				return fmt.Errorf("Error closing local plugin binary: %s", err)
			}
			return nil
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
			err := addrValidate(lbp.Addr)
			if err != nil {
				return "", errors.Wrap(err, "addr string validate")
			}
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

func addrValidate(addr string) error {
	compile := regexp.MustCompile("[0-9]{2,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:[0-9]{4,6}")
	match := compile.MatchString(addr)
	if !match {
		return errors.Errorf("fail to match addr [%s], please check addr is valide for [0-9]{2,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:[0-9]{4,6}", addr)
	}

	_, err := net.DialTimeout("tcp", addr, time.Second*3)
	if err != nil {
		return errors.Wrapf(err, "fail to dail addr [%s], please check addr", addr)
	}
	return nil
}

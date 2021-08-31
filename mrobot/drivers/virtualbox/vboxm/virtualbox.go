package virtualbox

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/ssh_cmd"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm/base"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm/state"
	mcnutils "github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm/util"
	"golang.org/x/crypto/ssh"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultCPU                 = 1
	defaultMemory              = 1024
	defaultBoot2DockerURL      = "https://github.com/boot2docker/boot2docker/releases/download/v19.03.12/boot2docker.iso"
	defaultBoot2DockerImportVM = ""
	defaultHostOnlyCIDR        = "192.168.99.1/24"
	defaultHostOnlyNictype     = "82540EM"
	defaultHostOnlyPromiscMode = "deny"
	defaultUIType              = "headless"
	defaultHostOnlyNoDHCP      = false
	defaultDiskSize            = 20000
	defaultDNSProxy            = true
	defaultDNSResolver         = true
	defaultShareFolder         = "/Users:Users"
)

var (
	ErrUnableToGenerateRandomIP = errors.New("unable to generate random IP")
	ErrMustEnableVTX            = errors.New("this computer doesn't have VT-X/AMD-v enabled. Enabling it in the BIOS is mandatory")
	ErrNotCompatibleWithHyperV  = errors.New("this computer is running Hyper-V. VirtualBox won't boot a 64bits VM when Hyper-V is activated. Either use Hyper-V as a driver, or disable the Hyper-V hypervisor. (To skip this check, use --virtualbox-no-vtx-check)")
	ErrNetworkAddrCidr          = errors.New("host-only cidr must be specified with a host address, not a network address")
	ErrNetworkAddrCollision     = errors.New("host-only cidr conflicts with the network address of a host interface")
)

type Driver struct {
	*base.BaseDriver
	VBoxManager
	HostInterfaces
	b2dUpdater          B2DUpdater
	sshKeyGenerator     SSHKeyGenerator
	diskCreator         DiskCreator
	logsReader          LogsReader
	ipWaiter            IPWaiter
	randomInter         RandomInter
	sleeper             Sleeper
	CPU                 int
	Memory              int
	DiskSize            int
	NatNicType          string
	Boot2DockerURL      string
	Boot2DockerImportVM string
	HostDNSResolver     bool
	HostOnlyCIDR        string
	HostOnlyNicType     string
	HostOnlyPromiscMode string
	UIType              string
	HostOnlyNoDHCP      bool
	NoShare             bool
	DNSProxy            bool
	NoVTXCheck          bool
	ShareFolder         string

	ctx            context.Context
	cli            *ssh.Client
	rmtHost        string
	localSSHKeyMap map[string]string
}

// NewRMTDriver creates a new VirtualBox remote driver with default settings.
func NewRMTDriver(ctx context.Context, hostName, storePath, rmtHost string, cli *ssh_cmd.SSHCli) *Driver {
	return &Driver{
		VBoxManager:         NewVBoxRemoteCmdManager(ctx, cli),
		b2dUpdater:          NewRMTB2DUpdater(ctx, cli.SSHCli, "/Users/wanghengfang/GolandProjects/cmapp/mrobot/bin/boot2docker.iso"),
		sshKeyGenerator:     NewRmtSSHKeyGenerator(ctx, cli.SSHCli),
		diskCreator:         NewFileDiskCreator(ctx, cli.SSHCli),
		logsReader:          NewRmtLogsReader(ctx, cli.SSHCli),
		ipWaiter:            NewIPWaiter(),
		randomInter:         NewRandomInter(),
		sleeper:             NewSleeper(),
		HostInterfaces:      NewHostInterfaces(),
		Memory:              defaultMemory,
		CPU:                 defaultCPU,
		DiskSize:            defaultDiskSize,
		NatNicType:          defaultHostOnlyNictype,
		HostOnlyCIDR:        defaultHostOnlyCIDR,
		HostOnlyNicType:     defaultHostOnlyNictype,
		HostOnlyPromiscMode: defaultHostOnlyPromiscMode,
		UIType:              defaultUIType,
		HostOnlyNoDHCP:      defaultHostOnlyNoDHCP,
		DNSProxy:            defaultDNSProxy,
		HostDNSResolver:     defaultDNSResolver,
		Boot2DockerURL:      defaultBoot2DockerURL,
		BaseDriver: &base.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
		ctx:            ctx,
		cli:            cli.SSHCli,
		rmtHost:        rmtHost,
		ShareFolder:    defaultShareFolder,
		localSSHKeyMap: map[string]string{},
	}
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []base.Flag {
	return []base.Flag{
		base.IntFlag{
			Name:   "virtualbox-memory",
			Usage:  "Size of memory for host in MB",
			Value:  defaultMemory,
			EnvVar: "VIRTUALBOX_MEMORY_SIZE",
		},
		base.IntFlag{
			Name:   "virtualbox-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  defaultCPU,
			EnvVar: "VIRTUALBOX_CPU_COUNT",
		},
		base.IntFlag{
			Name:   "virtualbox-disk-size",
			Usage:  "Size of disk for host in MB",
			Value:  defaultDiskSize,
			EnvVar: "VIRTUALBOX_DISK_SIZE",
		},
		base.StringFlag{
			Name:   "virtualbox-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  defaultBoot2DockerURL,
			EnvVar: "VIRTUALBOX_BOOT2DOCKER_URL",
		},
		base.StringFlag{
			Name:   "virtualbox-import-boot2docker-vm",
			Usage:  "The name of a Boot2Docker VM to import",
			Value:  defaultBoot2DockerImportVM,
			EnvVar: "VIRTUALBOX_BOOT2DOCKER_IMPORT_VM",
		},
		base.BoolFlag{
			Name:   "virtualbox-host-dns-resolver",
			Usage:  "Use the host DNS resolver",
			EnvVar: "VIRTUALBOX_HOST_DNS_RESOLVER",
		},
		base.StringFlag{
			Name:   "virtualbox-nat-nictype",
			Usage:  "Specify the Network Adapter Type",
			Value:  defaultHostOnlyNictype,
			EnvVar: "VIRTUALBOX_NAT_NICTYPE",
		},
		base.StringFlag{
			Name:   "virtualbox-hostonly-cidr",
			Usage:  "Specify the Host Only CIDR",
			Value:  defaultHostOnlyCIDR,
			EnvVar: "VIRTUALBOX_HOSTONLY_CIDR",
		},
		base.StringFlag{
			Name:   "virtualbox-hostonly-nictype",
			Usage:  "Specify the Host Only Network Adapter Type",
			Value:  defaultHostOnlyNictype,
			EnvVar: "VIRTUALBOX_HOSTONLY_NIC_TYPE",
		},
		base.StringFlag{
			Name:   "virtualbox-hostonly-nicpromisc",
			Usage:  "Specify the Host Only Network Adapter Promiscuous Mode",
			Value:  defaultHostOnlyPromiscMode,
			EnvVar: "VIRTUALBOX_HOSTONLY_NIC_PROMISC",
		},
		base.StringFlag{
			Name:   "virtualbox-ui-type",
			Usage:  "Specify the UI Type: (gui|sdl|headless|separate)",
			Value:  defaultUIType,
			EnvVar: "VIRTUALBOX_UI_TYPE",
		},
		base.BoolFlag{
			Name:   "virtualbox-hostonly-no-dhcp",
			Usage:  "Disable the Host Only DHCP Server",
			EnvVar: "VIRTUALBOX_HOSTONLY_NO_DHCP",
		},
		base.BoolFlag{
			Name:   "virtualbox-no-share",
			Usage:  "Disable the mount of your home directory",
			EnvVar: "VIRTUALBOX_NO_SHARE",
		},
		base.BoolFlag{
			Name:   "virtualbox-no-dns-proxy",
			Usage:  "Disable proxying all DNS requests to the host",
			EnvVar: "VIRTUALBOX_NO_DNS_PROXY",
		},
		base.BoolFlag{
			Name:   "virtualbox-no-vtx-check",
			Usage:  "Disable checking for the availability of hardware virtualization before the vm is started",
			EnvVar: "VIRTUALBOX_NO_VTX_CHECK",
		},
		base.StringFlag{
			EnvVar: "VIRTUALBOX_SHARE_FOLDER",
			Name:   "virtualbox-share-folder",
			Usage:  "Mount the specified directory instead of the default home location. Format: dir:name",
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.rmtHost, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "virtualbox"
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) SetConfigFromFlags(flags base.DriverOptions) error {
	d.CPU = flags.Int("virtualbox-cpu-count")
	d.Memory = flags.Int("virtualbox-memory")
	d.DiskSize = flags.Int("virtualbox-disk-size")
	d.Boot2DockerURL = flags.String("virtualbox-boot2docker-url")
	d.SetSwarmConfigFromFlags(flags)
	d.SSHUser = "docker"
	d.Boot2DockerImportVM = flags.String("virtualbox-import-boot2docker-vm")
	d.HostDNSResolver = flags.Bool("virtualbox-host-dns-resolver")
	d.NatNicType = flags.String("virtualbox-nat-nictype")
	d.HostOnlyCIDR = flags.String("virtualbox-hostonly-cidr")
	d.HostOnlyNicType = flags.String("virtualbox-hostonly-nictype")
	d.HostOnlyPromiscMode = flags.String("virtualbox-hostonly-nicpromisc")
	d.UIType = flags.String("virtualbox-ui-type")
	d.HostOnlyNoDHCP = flags.Bool("virtualbox-hostonly-no-dhcp")
	d.NoShare = flags.Bool("virtualbox-no-share")
	d.DNSProxy = !flags.Bool("virtualbox-no-dns-proxy")
	d.NoVTXCheck = flags.Bool("virtualbox-no-vtx-check")
	d.ShareFolder = flags.String("virtualbox-share-folder")

	return nil
}

// PreCreateCheck checks that VBoxManage exists and works
func (d *Driver) PreCreateCheck() error {
	// Check that VBoxManage exists and works
	version, err := d.vbmOut("--version")
	if err != nil {
		return err
	}

	// Check that VBoxManage is of a supported version
	if err = checkVBoxManageVersion(strings.TrimSpace(version)); err != nil {
		return err
	}

	if !d.NoVTXCheck {
		if isHyperVInstalled() {
			return ErrNotCompatibleWithHyperV
		}

		if d.IsVTXDisabled() {
			return ErrMustEnableVTX
		}
	}

	// Downloading boot2docker to cache should be done here to make sure
	// that a download failure will not leave a machine half created.
	if err := d.b2dUpdater.UpdateISOCache(d.StorePath, d.Boot2DockerURL); err != nil {
		return err
	}

	// Check that Host-only interfaces are ok
	if _, err = listHostOnlyAdapters(d.VBoxManager); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Create() error {
	if err := d.CreateVM(); err != nil {
		return err
	}

	log.Info(d.ctx, "Starting the VM...")
	return d.Start()
}

func (d *Driver) CreateVM() error {
	//  这里要实现拷贝iso到远程主机 --> 已完成
	if err := d.b2dUpdater.CopyIsoToMachineDir(d.StorePath, d.MachineName, d.Boot2DockerURL); err != nil {
		return err
	}

	log.Info(d.ctx, "Creating VirtualBox VM...")

	// import b2d VM if requested
	if d.Boot2DockerImportVM != "" {
		name := d.Boot2DockerImportVM

		// make sure vm is stopped
		_ = d.vbm("controlvm", name, "poweroff")

		diskInfo, err := getVMDiskInfo(name, d.VBoxManager)
		if err != nil {
			return err
		}

		if _, err := os.Stat(diskInfo.Path); err != nil {
			return err
		}

		if err := d.vbm("clonehd", diskInfo.Path, d.diskPath()); err != nil {
			return err
		}

		log.Debugf(d.ctx, "Importing VM settings...")
		vmInfo, err := getVMInfo(name, d.VBoxManager)
		if err != nil {
			return err
		}

		d.CPU = vmInfo.CPUs
		d.Memory = vmInfo.Memory

		log.Debugf(d.ctx, "Importing SSH key...")
		keyPath := filepath.Join(mcnutils.GetHomeDir(), ".ssh", "id_boot2docker")
		if err := mcnutils.CopyFile(keyPath, d.GetSSHKeyPath()); err != nil {
			return err
		}
	} else {
		log.Infof(d.ctx, "Creating SSH key...")
		sshPubKeyPath, err := d.sshKeyGenerator.Generate(d.GetSSHKeyPath())
		if err != nil {
			return err
		}
		if len(sshPubKeyPath) == 0 {
			sshPubKeyPath = d.publicSSHKeyPath()
		}
		d.localSSHKeyMap[d.MachineName] = filepath.Dir(sshPubKeyPath)
		d.SSHKeyPath = filepath.Join(filepath.Dir(sshPubKeyPath), "id_rsa")
		log.Debugf(d.ctx, "Creating disk image...")
		if err := d.diskCreator.Create(d.DiskSize, sshPubKeyPath, d.diskPath(), d); err != nil {
			return err
		}
	}

	if err := d.vbm("createvm",
		"--basefolder", d.ResolveStorePath("."),
		"--name", d.MachineName,
		"--register"); err != nil {
		return err
	}

	log.Debugf(d.ctx, "VM CPUS: %d", d.CPU)
	log.Debugf(d.ctx, "VM Memory: %d", d.Memory)

	cpus := d.CPU
	if cpus < 1 {
		cpus = int(runtime.NumCPU())
	}
	if cpus > 32 {
		cpus = 32
	}

	hostDNSResolver := "off"
	if d.HostDNSResolver {
		hostDNSResolver = "on"
	}

	dnsProxy := "off"
	if d.DNSProxy {
		dnsProxy = "on"
	}

	var modifyFlags = []string{
		"modifyvm", d.MachineName,
		"--firmware", "bios",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--bioslogodisplaytime", "0",
		"--biosbootmenu", "disabled",
		"--ostype", "Linux26_64",
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memory", fmt.Sprintf("%d", d.Memory),
		"--acpi", "on",
		"--ioapic", "on",
		"--rtcuseutc", "on",
		"--natdnshostresolver1", hostDNSResolver,
		"--natdnsproxy1", dnsProxy,
		"--cpuhotplug", "off",
		"--pae", "on",
		"--hpet", "on",
		"--hwvirtex", "on",
		"--nestedpaging", "on",
		"--largepages", "on",
		"--vtxvpid", "on",
		"--accelerate3d", "off",
		"--boot1", "dvd"}

	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		modifyFlags = append(modifyFlags, "--longmode", "on")
	}

	if err := d.vbm(modifyFlags...); err != nil {
		return err
	}

	if err := d.vbm("modifyvm", d.MachineName,
		"--nic1", "nat",
		"--nictype1", d.NatNicType,
		"--cableconnected1", "on"); err != nil {
		return err
	}

	if err := d.vbm("storagectl", d.MachineName,
		"--name", "SATA",
		"--add", "sata",
		"--hostiocache", "on"); err != nil {
		return err
	}

	if err := d.vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "0",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", d.ResolveStorePath("boot2docker.iso")); err != nil {
		return err
	}

	if err := d.vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "1",
		"--device", "0",
		"--type", "hdd",
		"--medium", d.diskPath()); err != nil {
		return err
	}

	// let VBoxService do nice magic automounting (when it's used)
	if err := d.vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/"); err != nil {
		return err
	}
	if err := d.vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountDir", "/"); err != nil {
		return err
	}

	shareName, shareDir := getShareDriveAndName()

	if d.ShareFolder != "" {
		shareDir, shareName = parseShareFolder(d.ShareFolder)
	}

	if shareDir != "" && !d.NoShare {
		log.Debugf(d.ctx, "setting up shareDir '%s' -> '%s'", shareDir, shareName)
		// remote check dir exist
		out, err := d.execCmd(fmt.Sprintf("ls %s", shareDir))
		if err != nil {
			return err
		}
		if strings.Contains(out, "No such file or directory") {
			return errors.Errorf("share dir [%s] not exist", shareDir)
		}
		if shareName == "" {
			// parts of the VBox internal code are buggy with share names that start with "/"
			shareName = strings.TrimLeft(shareDir, "/")
			// TODO do some basic Windows -> MSYS path conversion
			// ie, s!^([a-z]+):[/\\]+!\1/!; s!\\!/!g
		}

		// woo, shareDir exists!  let's carry on!
		if err := d.vbm("sharedfolder", "add", d.MachineName, "--name", shareName, "--hostpath", shareDir, "--automount"); err != nil {
			return err
		}

		// enable symlinks
		if err := d.vbm("setextradata", d.MachineName, "VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+shareName, "1"); err != nil {
			return err
		}
	}

	return nil
}

func parseShareFolder(shareFolder string) (string, string) {
	split := strings.Split(shareFolder, ":")
	shareDir := strings.Join(split[:len(split)-1], ":")
	shareName := split[len(split)-1]
	return shareDir, shareName
}

func (d *Driver) hostOnlyIPAvailable() bool {
	ip, err := d.GetIP()
	if err != nil {
		log.Debugf(d.ctx, "ERROR getting IP: %s", err)
		return false
	}
	if ip == "" {
		log.Debug(d.ctx, "Strangely, there was no error attempting to get the IP, but it was still empty.")
		return false
	}

	log.Debugf(d.ctx, "IP is %s", ip)
	return true
}

func (d *Driver) Start() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	var hostOnlyAdapter *hostOnlyNetwork
	if s == state.Stopped {
		log.Infof(d.ctx, "Check network to re-create if needed...")

		if hostOnlyAdapter, err = d.setupHostOnlyNetwork(d.MachineName); err != nil {
			return fmt.Errorf("Error setting up host only network on machine start: %s", err)
		}
	}

	switch s {
	case state.Stopped, state.Saved:
		d.SSHPort, err = rmtSetPortForwarding(d, 1, "ssh", "tcp", 22)
		if err != nil {
			return err
		}

		if err := d.vbm("startvm", d.MachineName, "--type", d.UIType); err != nil {
			if lines, readErr := d.readVBoxLog(); readErr == nil && len(lines) > 0 {
				return fmt.Errorf("Unable to start the VM: %s\nDetails: %s", err, lines[len(lines)-1])
			}
			return fmt.Errorf("Unable to start the VM: %s", err)
		}
	case state.Paused:
		if err := d.vbm("controlvm", d.MachineName, "resume", "--type", d.UIType); err != nil {
			return err
		}
		log.Infof(d.ctx, "Resuming VM ...")
	default:
		log.Infof(d.ctx, "VM not in restartable state")
	}

	if !d.NoVTXCheck {
		// Verify that VT-X is not disabled in the started VM
		vtxIsDisabled, err := d.IsVTXDisabledInTheVM()
		if err != nil {
			return fmt.Errorf("Checking if hardware virtualization is enabled failed: %s", err)
		}

		if vtxIsDisabled {
			return ErrMustEnableVTX
		}
	}

	log.Infof(d.ctx, "Waiting for an IP...")
	if err := d.ipWaiter.Wait(d); err != nil {
		log.Infof(d.ctx, "ip wait err [%v]", err)
		return err
	}

	if hostOnlyAdapter == nil {
		return nil
	}

	// Check that the host-only adapter we just created can still be found
	// Sometimes it is corrupted after the VM is started.
	nets, err := listHostOnlyAdapters(d.VBoxManager)
	if err != nil {
		return err
	}

	ip, network, err := parseAndValidateCIDR(d.HostOnlyCIDR)
	if err != nil {
		return err
	}

	err = validateNoIPCollisions(d.HostInterfaces, network, nets)
	if err != nil {
		return err
	}

	hostOnlyNet := getHostOnlyAdapter(d.ctx, nets, ip, network.Mask)
	if hostOnlyNet != nil {
		// OK, we found a valid host-only adapter
		return nil
	}

	// This happens a lot on windows. The adapter has an invalid IP and the VM has the same IP
	log.Warn(d.ctx, "The host-only adapter is corrupted. Let's stop the VM, fix the host-only adapter and restart the VM")
	if err := d.Stop(); err != nil {
		return err
	}

	// We have to be sure the host-only adapter is not used by the VM
	d.sleeper.Sleep(5 * time.Second)

	log.Debugf(d.ctx, "Fixing %+v...", hostOnlyAdapter)
	if err := hostOnlyAdapter.SaveIPv4(d.VBoxManager); err != nil {
		return err
	}

	// We have to be sure the adapter is updated before starting the VM
	d.sleeper.Sleep(5 * time.Second)

	if err := d.vbm("startvm", d.MachineName, "--type", d.UIType); err != nil {
		return fmt.Errorf("Unable to start the VM: %s", err)
	}

	log.Infof(d.ctx, "Waiting for an IP...")
	return d.ipWaiter.Wait(d)
}

func (d *Driver) Stop() error {
	currentState, err := d.GetState()
	if err != nil {
		return err
	}

	if currentState == state.Paused {
		if err := d.vbm("controlvm", d.MachineName, "resume"); err != nil { // , "--type", d.UIType
			return err
		}
		log.Infof(d.ctx, "Resuming VM ...")
	}

	if err := d.vbm("controlvm", d.MachineName, "acpipowerbutton"); err != nil {
		return err
	}
	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}
		if s == state.Running {
			d.sleeper.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	d.IPAddress = ""

	return nil
}

// Restart restarts a machine which is known to be running.
func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return fmt.Errorf("Problem stopping the VM: %s", err)
	}

	if err := d.Start(); err != nil {
		return fmt.Errorf("Problem starting the VM: %s", err)
	}

	d.IPAddress = ""

	return d.ipWaiter.Wait(d)
}

func (d *Driver) Kill() error {
	return d.vbm("controlvm", d.MachineName, "poweroff")
}

func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err == ErrMachineNotExist {
		return nil
	}
	if err != nil {
		return err
	}

	if s != state.Stopped && s != state.Saved {
		if err := d.Kill(); err != nil {
			return err
		}
	}

	return d.vbm("unregistervm", "--delete", d.MachineName)
}

func (d *Driver) GetState() (state.State, error) {
	stdout, stderr, err := d.vbmOutErr("showvminfo", d.MachineName, "--machinereadable")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return state.Error, ErrMachineNotExist
		}
		return state.Error, err
	}
	re := regexp.MustCompile(`(?m)^VMState="(\w+)"`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 1 {
		return state.None, nil
	}
	switch groups[1] {
	case "running":
		return state.Running, nil
	case "paused":
		return state.Paused, nil
	case "saved":
		return state.Saved, nil
	case "poweroff", "aborted":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) getHostOnlyMACAddress() (string, error) {
	// Return the MAC address of the host-only adapter
	// assigned to this machine. The returned address
	// is lower-cased and does not contain colons.

	stdout, stderr, err := d.vbmOutErr("showvminfo", d.MachineName, "--machinereadable")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return "", ErrMachineNotExist
		}
		return "", err
	}

	// First, we get the number of the host-only interface
	re := regexp.MustCompile(`(?m)^hostonlyadapter([\d]+)`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 2 {
		return "", errors.New("Machine does not have a host-only adapter")
	}

	// Then we grab the MAC address based on that number
	adapterNumber := groups[1]
	re = regexp.MustCompile(fmt.Sprintf("(?m)^macaddress%s=\"(.*)\"", adapterNumber))
	groups = re.FindStringSubmatch(stdout)
	if len(groups) < 2 {
		return "", fmt.Errorf("Could not find MAC address for adapter %v", adapterNumber)
	}

	return strings.ToLower(groups[1]), nil
}

func (d *Driver) parseIPForMACFromIPAddr(ipAddrOutput string, macAddress string) (string, error) {
	// Given the output of "ip addr show" on the VM, return the IPv4 address
	// of the interface with the given MAC address.

	lines := strings.Split(ipAddrOutput, "\n")
	returnNextIP := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "link") { // line contains MAC address
			vals := strings.Split(line, " ")
			if len(vals) >= 2 {
				macBlock := vals[1]
				macWithoutColons := strings.Replace(macBlock, ":", "", -1)
				if macWithoutColons == macAddress { // we are in the correct device block
					returnNextIP = true
				}
			}
		} else if strings.HasPrefix(line, "inet") && !strings.HasPrefix(line, "inet6") && returnNextIP {
			vals := strings.Split(line, " ")
			if len(vals) >= 2 {
				return vals[1][:strings.Index(vals[1], "/")], nil
			}
		}
	}

	return "", fmt.Errorf("Could not find matching IP for MAC address %v", macAddress)
}

func (d *Driver) GetIP() (string, error) {
	// DHCP is used to get the IP, so virtualbox hosts don't have IPs unless
	// they are running
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", base.ErrHostIsNotRunning
	}

	macAddress, err := d.getHostOnlyMACAddress()
	if err != nil {
		return "", err
	}

	log.Debugf(d.ctx, "Host-only MAC: %s\n", macAddress)

	output, err := base.RunSSHCommandFromDriver(d, "ip addr show")
	if err != nil {
		return "", err
	}

	log.Debugf(d.ctx, "SSH returned: %s\nEND SSH\n", output)

	ipAddress, err := d.parseIPForMACFromIPAddr(output, macAddress)
	if err != nil {
		return "", err
	}

	return ipAddress, nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return d.ResolveStorePath("disk.vmdk")
}

func (d *Driver) setupHostOnlyNetwork(machineName string) (*hostOnlyNetwork, error) {
	hostOnlyCIDR := d.HostOnlyCIDR

	// This is to assist in migrating from version 0.2 to 0.3 format
	// it should be removed in a later release
	if hostOnlyCIDR == "" {
		hostOnlyCIDR = defaultHostOnlyCIDR
	}

	ip, network, err := parseAndValidateCIDR(hostOnlyCIDR)
	if err != nil {
		return nil, err
	}

	nets, err := listHostOnlyAdapters(d.VBoxManager)
	if err != nil {
		return nil, err
	}

	// TODO: remote validate no ip collisions, may be implement 'HostInterfaces' adapter
	//err = validateNoIPCollisions(d.HostInterfaces, network, nets)
	//if err != nil {
	//	return nil, err
	//}

	log.Debugf(d.ctx, "Searching for hostonly interface for IPv4: %s and Mask: %s", ip, network.Mask)
	hostOnlyAdapter, err := getOrCreateHostOnlyNetwork(d.ctx, ip, network.Mask, nets, d.VBoxManager)
	if err != nil {
		return nil, err
	}

	if err := removeOrphanDHCPServers(d.ctx, d.VBoxManager); err != nil {
		return nil, err
	}

	dhcpAddr, err := getRandomIPinSubnet(d, ip)
	if err != nil {
		return nil, err
	}

	lowerIP, upperIP := getDHCPAddressRange(dhcpAddr, network)

	log.Debugf(d.ctx, "Adding/Modifying DHCP server %q with address range %q - %q...", dhcpAddr, lowerIP, upperIP)

	dhcp := dhcpServer{}
	dhcp.IPv4.IP = dhcpAddr
	dhcp.IPv4.Mask = network.Mask
	dhcp.LowerIP = lowerIP
	dhcp.UpperIP = upperIP
	dhcp.Enabled = !d.HostOnlyNoDHCP
	if err := addHostOnlyDHCPServer(d.ctx, hostOnlyAdapter.Name, dhcp, d.VBoxManager); err != nil {
		return nil, err
	}

	if err := d.vbm("modifyvm", machineName,
		"--nic2", "hostonly",
		"--nictype2", d.HostOnlyNicType,
		"--nicpromisc2", d.HostOnlyPromiscMode,
		"--hostonlyadapter2", hostOnlyAdapter.Name,
		"--cableconnected2", "on"); err != nil {
		return nil, err
	}

	return hostOnlyAdapter, nil
}

func getDHCPAddressRange(dhcpAddr net.IP, network *net.IPNet) (lowerIP net.IP, upperIP net.IP) {
	nAddr := network.IP.To4()
	ones, bits := network.Mask.Size()

	if ones <= 24 {
		// For a /24 subnet, use the original behavior of allowing the address range
		// between x.x.x.100 and x.x.x.254.
		lowerIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(100))
		upperIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(254))
		return
	}

	// Start the lowerIP range one address above the selected DHCP address.
	lowerIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], dhcpAddr.To4()[3]+1)

	// The highest D-part of the address A.B.C.D in this subnet is at 2^n - 1,
	// where n is the number of available bits in the subnet. Since the highest
	// address is reserved for subnet broadcast, the highest *assignable* address
	// is at (2^n - 1) - 1 == 2^n - 2.
	maxAssignableSubnetAddress := (byte)((1 << (uint)(bits-ones)) - 2)
	upperIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], maxAssignableSubnetAddress)
	return
}

func parseAndValidateCIDR(hostOnlyCIDR string) (net.IP, *net.IPNet, error) {
	ip, network, err := net.ParseCIDR(hostOnlyCIDR)
	if err != nil {
		return nil, nil, err
	}

	networkAddress := network.IP.To4()
	if ip.Equal(networkAddress) {
		return nil, nil, ErrNetworkAddrCidr
	}

	return ip, network, nil
}

// validateNoIPCollisions ensures no conflicts between the host's network interfaces and the vbox host-only network that
// will be used for machine vm instances.
func validateNoIPCollisions(hif HostInterfaces, hostOnlyNet *net.IPNet, currHostOnlyNets map[string]*hostOnlyNetwork) error {
	hostOnlyByCIDR := map[string]*hostOnlyNetwork{}
	//listHostOnlyAdapters returns a map w/ virtualbox net names as key.  Rekey to CIDRs
	for _, n := range currHostOnlyNets {
		ipnet := net.IPNet{IP: n.IPv4.IP, Mask: n.IPv4.Mask}
		hostOnlyByCIDR[ipnet.String()] = n
	}

	m, err := listHostInterfaces(hif, hostOnlyByCIDR)
	if err != nil {
		return err
	}

	collision, err := checkIPNetCollision(hostOnlyNet, m)
	if err != nil {
		return err
	}

	if collision {
		return ErrNetworkAddrCollision
	}
	return nil
}

func getRmtAvailableTCPPort(ctx context.Context, rmtHost string) (int, error) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 20; i++ {
		p := randInt(30001, 65534)
		cn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", rmtHost, p), 3*time.Second)
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				log.Infof(ctx, "remote host [%s] port [%d] not in use", rmtHost, p)
				resultPort := p
				return resultPort, nil
			}
		}
		if cn != nil {
			_ = cn.Close()
		}
	}
	return -1, fmt.Errorf("unable to allocate tcp port")
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// Select an available port, trying the specified
// port first, falling back on an OS selected port.
func getAvailableTCPPort(port int) (int, error) {
	for i := 0; i <= 10; i++ {
		ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return 0, err
		}
		defer ln.Close()
		addr := ln.Addr().String()
		addrParts := strings.SplitN(addr, ":", 2)
		p, err := strconv.Atoi(addrParts[1])
		if err != nil {
			return 0, err
		}
		if p != 0 {
			port = p
			return port, nil
		}
		port = 0 // Throw away the port hint before trying again
		time.Sleep(1)
	}
	return 0, fmt.Errorf("unable to allocate tcp port")
}

// Setup a NAT port forwarding entry.
func setPortForwarding(d *Driver, interfaceNum int, mapName, protocol string, guestPort, desiredHostPort int) (int, error) {
	actualHostPort, err := getAvailableTCPPort(desiredHostPort)
	if err != nil {
		return -1, err
	}
	if desiredHostPort != actualHostPort && desiredHostPort != 0 {
		log.Debugf(d.ctx, "NAT forwarding host port for guest port %d (%s) changed from %d to %d",
			guestPort, mapName, desiredHostPort, actualHostPort)
	}
	cmd := fmt.Sprintf("--natpf%d", interfaceNum)
	d.vbm("modifyvm", d.MachineName, cmd, "delete", mapName)
	if err := d.vbm("modifyvm", d.MachineName,
		cmd, fmt.Sprintf("%s,%s,127.0.0.1,%d,,%d", mapName, protocol, actualHostPort, guestPort)); err != nil {
		return -1, err
	}
	return actualHostPort, nil
}

// rmtSetPortForwarding Setup a NAT port forwarding entry.
func rmtSetPortForwarding(d *Driver, interfaceNum int, mapName, protocol string, guestPort int) (int, error) {
	actualHostPort, err := getRmtAvailableTCPPort(d.ctx, d.rmtHost)
	if err != nil {
		return -1, err
	}
	if actualHostPort != -1 {
		log.Debugf(d.ctx, "NAT forwarding host port for guest port %d (%s) changed from %d to %d",
			guestPort, mapName, 0, actualHostPort)
	}
	cmd := fmt.Sprintf("--natpf%d", interfaceNum)
	d.vbm("modifyvm", d.MachineName, cmd, "delete", mapName)
	if err := d.vbm("modifyvm", d.MachineName,
		cmd, fmt.Sprintf("%s,%s,0.0.0.0,%d,,%d", mapName, protocol, actualHostPort, guestPort)); err != nil {
		return -1, err
	}
	return actualHostPort, nil
}

// getRandomIPinSubnet returns a pseudo-random net.IP in the same
// subnet as the IP passed
func getRandomIPinSubnet(d *Driver, baseIP net.IP) (net.IP, error) {
	var dhcpAddr net.IP

	nAddr := baseIP.To4()
	// select pseudo-random DHCP addr; make sure not to clash with the host
	// only try 5 times and bail if no random received
	for i := 0; i < 5; i++ {
		n := d.randomInter.RandomInt(24) + 1
		if byte(n) != nAddr[3] {
			dhcpAddr = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(n))
			break
		}
	}

	if dhcpAddr == nil {
		return nil, ErrUnableToGenerateRandomIP
	}

	return dhcpAddr, nil
}

func detectVBoxManageCmdInPath() string {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}

func (d *Driver) readVBoxLog() ([]string, error) {
	logPath := filepath.Join(d.ResolveStorePath(d.MachineName), "Logs", "VBox.log")
	log.Debugf(d.ctx, "Checking vm logs: %s", logPath)

	return d.logsReader.Read(logPath)
}

func (d *Driver) execCmd(cmd string) (string, error) {
	sess, err := d.cli.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "new session with ssh server")
	}
	defer sess.Close()

	// exe command
	res, err := sess.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "exec command [%s]", cmd)
	}
	return string(res), nil
}

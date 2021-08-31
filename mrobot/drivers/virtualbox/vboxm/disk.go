package virtualbox

import (
	"context"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	mcnutils "github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm/util"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type VirtualDisk struct {
	UUID string
	Path string
}

type DiskCreator interface {
	Create(size int, publicSSHKeyPath, diskPath string, vbm ...VBoxManager) error
}

func NewFileDiskCreator(ctx context.Context, cli *ssh.Client) DiskCreator {
	return &fileRmtDiskCreator{
		ctx: ctx,
		cli: cli,
	}
}

type fileRmtDiskCreator struct {
	ctx context.Context
	cli *ssh.Client
}

// Create Make a boot2docker VM disk image.
func (f *fileRmtDiskCreator) Create(size int, publicSSHKeyPath, diskPath string, vbms ...VBoxManager) error {
	log.Infof(f.ctx, "Creating %d MB hard disk image...", size)
	if len(vbms) == 0 {
		return errors.New("no vb manager found")
	}
	log.Infof(f.ctx, "public ssh key path [%s]", publicSSHKeyPath)
	tarBuf, err := mcnutils.MakeDiskImage(publicSSHKeyPath)
	if err != nil {
		return err
	}
	log.Debug(f.ctx, "Write tmp disk image")

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fs, err := ioutil.TempFile(dir, "diskvm.tmp")
	if err != nil {
		return err
	}

	defer func() {
		if err := removeFileIfExists(fs.Name()); err != nil {
			log.Warnf(context.Background(), "Error removing file: %s", err)
		}
	}()

	_, err = io.Copy(fs, tarBuf)
	if err != nil {
		return err
	}
	if err = fs.Close(); err != nil {
		return err
	}

	log.Infof(f.ctx, "fs name : %s", fs.Name())
	fdir := filepath.Dir(diskPath)

	out, err := f.execCmd(fmt.Sprintf("mkdir -p %s", fdir))
	if err != nil {
		return errors.Wrapf(err, "exec create dir [%s] command", fdir)
	}
	if len(out) != 0 {
		return errors.Errorf("fail to exec create dir [%s] command, res [%s]", fdir, out)
	}

	err = f.uploadFile(fs.Name(), filepath.Join(fdir, "disk.tar"))
	if err != nil {
		return errors.Wrap(err, "upload disk image")
	}
	sizeBytes := int64(size) << 20 // usually won't fit in 32-bit int (max 2GB)
	cmd := fmt.Sprintf("cat %s | /usr/local/bin/VBoxManage convertfromraw stdin %s %d --format VMDK",
		filepath.Join(fdir, "disk.tar"), diskPath, sizeBytes)
	out, err = f.execCmd(cmd)
	if err != nil {
		return errors.Wrap(err, "exec cmd")
	}
	log.Infof(f.ctx, "Currently exec convertfromraw res [%s]", out)
	time.Sleep(time.Second*5)
	return nil
}

func (f *fileRmtDiskCreator) uploadFile(file, rmtFile string) error {
	client, err := scp.NewClientBySSH(f.cli)
	if err != nil {
		log.Errorf(f.ctx, "Couldn't establish a connection to the remote server, Err: [%v]", err)
		return errors.Wrap(err, "establish connection to remote server")
	}
	// Open a file
	fs, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "open boot2docker.iso")
	}

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer fs.Close()

	// make sure dir is created
	rdir := filepath.Dir(rmtFile)
	out, err := f.execCmd(fmt.Sprintf("mkdir -p %s", rdir))
	if err != nil {
		return errors.Wrapf(err, "exec create dir [%s] command", rdir)
	}
	if len(out) != 0 {
		return errors.Errorf("fail to exec create dir [%s] command, res [%s]", rdir, out)
	}

	err = client.CopyFile(fs, rmtFile, "0777")
	if err != nil {
		return errors.Wrap(err, "copying file to remote")
	}
	return nil
}

func (f *fileRmtDiskCreator) execCmd(cmd string) (string, error) {
	sess, err := f.cli.NewSession()
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

func removeFileIfExists(name string) error {
	if _, err := os.Stat(name); err == nil {
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("Error removing temporary download file: %s", err)
		}
	}
	return nil
}

func NewDiskCreator(ctx context.Context) DiskCreator {
	return &defaultDiskCreator{ctx: ctx}
}

type defaultDiskCreator struct{ ctx context.Context }

// Create Make a boot2docker VM disk image.
func (c *defaultDiskCreator) Create(size int, publicSSHKeyPath, diskPath string, vbm ...VBoxManager) error {
	log.Debugf(c.ctx, "Creating %d MB hard disk image...", size)

	tarBuf, err := mcnutils.MakeDiskImage(publicSSHKeyPath)
	if err != nil {
		return err
	}

	log.Debug(c.ctx, "Calling inner createDiskImage")

	return createDiskImage(c.ctx, diskPath, size, tarBuf)
}

// createDiskImage makes a disk image at dest with the given size in MB. If r is
// not nil, it will be read as a raw disk image to convert from.
func createDiskImage(ctx context.Context, dest string, size int, r io.Reader) error {
	// Convert a raw image from stdin to the dest VMDK image.
	sizeBytes := int64(size) << 20 // usually won't fit in 32-bit int (max 2GB)
	// FIXME: why isn't this just using the vbm*() functions?
	cmd := exec.Command(vboxManageCmd, "convertfromraw", "stdin", dest,
		fmt.Sprintf("%d", sizeBytes), "--format", "VMDK")

	log.Debugf(ctx, "%v", cmd)

	if os.Getenv("MACHINE_DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	log.Debug(ctx, "Starting command")

	if err := cmd.Start(); err != nil {
		return err
	}

	log.Debug(ctx, "Copying to stdin")

	n, err := io.Copy(stdin, r)
	if err != nil {
		return err
	}

	log.Debug(ctx, "Filling zeroes")

	// The total number of bytes written to stdin must match sizeBytes, or
	// VBoxManage.exe on Windows will fail. Fill remaining with zeros.
	if left := sizeBytes - n; left > 0 {
		if err := zeroFill(stdin, left); err != nil {
			return err
		}
	}

	log.Debug(ctx, "Closing STDIN")

	// cmd won't exit until the stdin is closed.
	if err := stdin.Close(); err != nil {
		return err
	}

	log.Debug(ctx, "Waiting on cmd")

	return cmd.Wait()
}

// zeroFill writes n zero bytes into w.
func zeroFill(w io.Writer, n int64) error {
	const blocksize = 32 << 10
	zeros := make([]byte, blocksize)
	var k int
	var err error
	for n > 0 {
		if n > blocksize {
			k, err = w.Write(zeros)
		} else {
			k, err = w.Write(zeros[:n])
		}
		if err != nil {
			return err
		}
		n -= int64(k)
	}
	return nil
}

func getVMDiskInfo(name string, vbox VBoxManager) (*VirtualDisk, error) {
	out, err := vbox.vbmOut("showvminfo", name, "--machinereadable")
	if err != nil {
		return nil, err
	}

	disk := &VirtualDisk{}

	err = parseKeyValues(out, reEqualQuoteLine, func(key, val string) error {
		switch key {
		case "SATA-1-0":
			disk.Path = val
		case "SATA-ImageUUID-1-0":
			disk.UUID = val
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return disk, nil
}

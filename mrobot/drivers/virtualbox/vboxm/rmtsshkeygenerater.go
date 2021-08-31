package virtualbox

import (
	"context"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
)


func NewRmtSSHKeyGenerator(ctx context.Context, cli *ssh2.Client) SSHKeyGenerator {
	return &rmtSSHKeyGenerator{
		ctx: ctx,
		cli: cli,
	}
}

type rmtSSHKeyGenerator struct {
	ctx context.Context
	cli *ssh2.Client
}

func (r *rmtSSHKeyGenerator) Generate(path string) (string, error) {
	kp, err := ssh.NewKeyPair()
	if err != nil {
		return "", fmt.Errorf("fail to generating key pair: %v", err)
	}
	tmp := uuid.New().String()
	sshtmpdir := filepath.Join("sshkey", tmp)
	if _, err = os.Stat(sshtmpdir); errors.Is(err, os.ErrNotExist) {
		_ = os.MkdirAll(sshtmpdir, os.ModePerm)
	}
	privateKeyPath := fmt.Sprintf("%s/id_rsa", sshtmpdir)
	publicKeyPath := fmt.Sprintf("%s/id_rsa.pub", sshtmpdir)
	err = kp.WriteToFile(privateKeyPath, publicKeyPath)
	if err != nil {
		return "", errors.Wrapf(err, "write ssh key to path [%s]", sshtmpdir)
	}
	err = r.uploadFile(privateKeyPath, path)
	if err != nil {
		return "", errors.Wrapf(err, "write ssh private key to path [%s]", path)
	}
	err = r.uploadFile(publicKeyPath, fmt.Sprintf("%s.pub", path))
	if err != nil {
		return "", errors.Wrapf(err, "write ssh public key to path [%s.pub]", path)
	}

	return publicKeyPath, nil
}

func (r *rmtSSHKeyGenerator) uploadFile(file, rmtFile string) error {
	client, err := scp.NewClientBySSH(r.cli)
	if err != nil {
		log.Errorf(r.ctx, "Couldn't establish a connection to the remote server, Err: [%v]", err)
		return errors.Wrap(err, "establish connection to remote server")
	}
	// Open a file
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "open boot2docker.iso")
	}

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer f.Close()

	// make sure dir is created
	rdir := filepath.Dir(rmtFile)
	out, err := r.execCmd(fmt.Sprintf("mkdir -p %s", rdir))
	if err != nil {
		return errors.Wrapf(err, "exec create dir [%s] command", rdir)
	}
	if len(out) != 0 {
		return errors.Errorf("fail to exec create dir [%s] command, res [%s]", rdir, out)
	}

	err = client.CopyFile(f, rmtFile, "0777")
	if err != nil {
		return errors.Wrap(err, "copying file to remote")
	}
	return nil
}

func (r *rmtSSHKeyGenerator) execCmd(cmd string) (string, error) {
	sess, err := r.cli.NewSession()
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

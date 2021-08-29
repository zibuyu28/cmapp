package base

import (
	"context"
	"fmt"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm/ssh"
	mcnutils "github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm/util"
)

func GetSSHClientFromDriver(d Driver) (ssh.Client, error) {
	address, err := d.GetSSHHostname()
	if err != nil {
		return nil, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return nil, err
	}

	var auth *ssh.Auth
	if d.GetSSHKeyPath() == "" {
		auth = &ssh.Auth{}
	} else {
		auth = &ssh.Auth{
			Keys: []string{d.GetSSHKeyPath()},
		}
	}

	client, err := ssh.NewClient(context.Background(), d.GetSSHUsername(), address, port, auth)
	return client, err

}

func RunSSHCommandFromDriver(d Driver, command string) (string, error) {
	client, err := GetSSHClientFromDriver(d)
	if err != nil {
		return "", err
	}

	log.Debugf(context.Background(), "About to run SSH command:%s", command)

	output, err := client.Output(command)
	log.Debugf(context.Background(), "SSH cmd err, output: %v: %s", err, output)
	if err != nil {
		return "", fmt.Errorf(`ssh command error:
command : %s
err     : %v
output  : %s`, command, err, output)
	}

	return output, nil
}

func sshAvailableFunc(d Driver) func() bool {
	return func() bool {
		log.Debug(context.Background(), "Getting to WaitForSSH function...")
		if _, err := RunSSHCommandFromDriver(d, "exit 0"); err != nil {
			log.Debugf(context.Background(), "Error getting ssh command 'exit 0' : %s", err)
			return false
		}
		return true
	}
}

func WaitForSSH(d Driver) error {
	// Try to dial SSH for 30 seconds before timing out.
	if err := mcnutils.WaitFor(sshAvailableFunc(d)); err != nil {
		return fmt.Errorf("Too many retries waiting for SSH to be available.  Last error: %s", err)
	}
	return nil
}

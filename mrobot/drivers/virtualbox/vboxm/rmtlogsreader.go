package virtualbox

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"strings"
)

func NewRmtLogsReader(ctx context.Context, cli *ssh.Client) LogsReader {
	return &rmtVBoxLogsReader{
		ctx: ctx,
		cli: cli,
	}
}

type rmtVBoxLogsReader struct {
	ctx context.Context
	cli *ssh.Client
}

func (r *rmtVBoxLogsReader) Read(path string) ([]string, error) {

	cmd := fmt.Sprintf("cat %s", path)
	res, err := r.execCmd(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "exec remote cmd")
	}
	lines := strings.Split(res, "\n")
	return lines, nil
}

func (r *rmtVBoxLogsReader) execCmd(cmd string) (string, error) {
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

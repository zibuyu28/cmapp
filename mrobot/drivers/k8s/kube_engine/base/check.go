package kubeclient

import (
	"github.com/pkg/errors"
)

// CheckEnv check env
func (c *Client) CheckEnv() error {
	_, err := c.k.Discovery().ServerVersion()
	if err != nil {
		return errors.Wrap(err, "discover server version")
	}
	return nil
}

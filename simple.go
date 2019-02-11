package emotiva

import (
	"github.com/urfave/cli"
)

func Ping(c *cli.Context) error {
	ec, err := NewEmotivaController(c.GlobalString("address"))
	if err != nil {
		return err
	}

	_, err = ec.Ping()
	if err != nil {
		return err
	}

	ec.Close()

	return nil
}

func Control(c *cli.Context, command string) error {
	ec, err := NewEmotivaController(c.GlobalString("address"))
	if err != nil {
		return err
	}

	_, err = ec.Control(command, nil)
	if err != nil {
		return err
	}

	ec.Close()

	return nil
}

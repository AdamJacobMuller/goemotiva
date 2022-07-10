package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/adamjacobmuller/goemotiva"
	"github.com/adamjacobmuller/goemotiva/server"
	"github.com/urfave/cli"
)

func inputFuncGenerator(input string) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		return emotiva.Control(c, fmt.Sprintf(`<%s value="0" ack="yes" />`, input))
	}
}

func inputGenerator() []cli.Command {
	var inputs []cli.Command

	for _, input := range []string{
		"hdmi1",
		"hdmi2",
		"hdmi3",
		"hdmi4",
		"hdmi5",
		"hdmi6",
		"hdmi7",
		"hdmi8",
		"coax1",
		"coax2",
		"coax3",
		"coax4",
		"optical1",
		"optical2",
		"optical3",
		"optical4",
		"ARC",
	} {
		inputs = append(inputs, cli.Command{
			Name:   input,
			Usage:  "select input " + input,
			Action: inputFuncGenerator(input),
		})

	}

	return inputs
}

func menuJog(c *cli.Context) error {
	ec, err := emotiva.NewEmotivaController(c.GlobalString("address"))
	if err != nil {
		return err
	}

	_, err = ec.Control(`<menu value="0" ack="yes" />`, nil)
	if err != nil {
		return err
	}

	cc := make(chan os.Signal)
	signal.Notify(cc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cc
		_, err = ec.Control(`<menu value="0" ack="yes" />`, nil)
		if err != nil {
			panic(err)
		}
		exec.Command("stty", "-f", "/dev/tty", "echo").Run()
		os.Exit(0)
	}()

	//no buffering
	exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1").Run()
	//no visible output
	exec.Command("stty", "-f", "/dev/tty", "-echo").Run()

	reader := bufio.NewReader(os.Stdin)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}
		err = nil
		switch int(b) {
		case 27:
		case 91:
			continue
		case 65:
			// up
			_, err = ec.Control(`<up value="0" ack="yes" />`, nil)
		case 66:
			// down
			_, err = ec.Control(`<down value="0" ack="yes" />`, nil)
		case 67:
			// right
			_, err = ec.Control(`<right value="0" ack="yes" />`, nil)
		case 68:
			// left
			_, err = ec.Control(`<left value="0" ack="yes" />`, nil)
		case 10:
			// enter
			_, err = ec.Control(`<enter value="0" ack="yes" />`, nil)
		case 113:
			// q
			close(cc)
		default:
			fmt.Printf("unhandled int(b) is %d\n", int(b))
		}
		if err != nil {
			return err
		}
	}
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "address",
			Value:  "10.0.8.41",
			Usage:  "Emotiva address",
			EnvVar: "EMOTIVA_ADDRESS",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "volume",
			Usage: "set volume",
			Flags: []cli.Flag{
				cli.IntFlag{Name: "volume, v"},
			},
			Action: func(c *cli.Context) error {
				return emotiva.Control(c, fmt.Sprintf(`<set_volume value="%d" ack="yes" />`, c.Int("volume")))
			},
		},
		{
			Name:  "status",
			Usage: "status",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "parameters",
					Value: "power,source,volume,mode,audio_input,audio_bitstream,audio_bits,video_input,video_format,video_space",
				},
			},
			Action: func(c *cli.Context) error {
				parameters := strings.Split(c.String("parameters"), ",")
				return emotiva.Status(c, parameters)
			},
		},
		{
			Name:  "server",
			Usage: "server",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				_, err := server.NewServer(c.GlobalString("address"))
				if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "menu",
			Usage: "toggle menu",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return emotiva.Control(c, `<menu value="0" ack="yes" />`)
			},
			Subcommands: []cli.Command{
				{
					Name:   "jog",
					Usage:  "toggle menu and read input for keyboard navigation",
					Action: menuJog,
				},
			},
		},
		{
			Name:  "info",
			Usage: "display info screen",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return emotiva.Control(c, `<info value="0" ack="yes" />`)
			},
		},
		{
			Name:  "ping",
			Usage: "emotivaPing",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return emotiva.Ping(c)
			},
		},
		{
			Name:  "power",
			Usage: "set power state",
			Flags: []cli.Flag{},
			Subcommands: []cli.Command{
				{
					Name:  "on",
					Usage: "turn power on",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<power_on value="0" ack="yes" />`)
					},
				},
				{
					Name:  "off",
					Usage: "turn power off",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<power_off value="0" ack="yes" />`)
					},
				},
				{
					Name:  "standby",
					Usage: "set standby",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<standby value="0" ack="yes" />`)
					},
				},
			},
		},
		{
			Name:        "input",
			Usage:       "set input",
			Flags:       []cli.Flag{},
			Subcommands: inputGenerator(),
		},
		{
			Name:  "power",
			Usage: "set power state",
			Flags: []cli.Flag{},
			Subcommands: []cli.Command{
				{
					Name:  "on",
					Usage: "turn power on",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<power_on value="0" ack="yes" />`)
					},
				},
				{
					Name:  "off",
					Usage: "turn power off",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<power_off value="0" ack="yes" />`)
					},
				},
				{
					Name:  "standby",
					Usage: "set standby",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<standby value="0" ack="yes" />`)
					},
				},
			},
		},
		{
			Name:  "loudness",
			Usage: "set loudness",
			Flags: []cli.Flag{},
			Subcommands: []cli.Command{
				{
					Name:  "on",
					Usage: "turn loudness on",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<loudness_on value="0" ack="yes" />`)
					},
				},
				{
					Name:  "off",
					Usage: "turn loudness off",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<loudness_off value="0" ack="yes" />`)
					},
				},
				{
					Name:  "toggle",
					Usage: "toggle loudness",
					Action: func(c *cli.Context) error {
						return emotiva.Control(c, `<loudness value="0" ack="yes" />`)
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

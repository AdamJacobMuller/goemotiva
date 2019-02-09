package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

func emotivaRW(c *cli.Context, command, wPort, rPort string, target interface{}) (string, error) {
	listenAddr, err := net.ResolveUDPAddr("udp", ":"+rPort)
	if err != nil {
		return "", err
	}

	listenConn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return "", err
	}

	conn, err := net.Dial("udp", c.GlobalString("address")+":"+wPort)
	if err != nil {
		return "", err
	}

	data := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>%s", command)

	n, err := conn.Write([]byte(data))
	if err != nil {
		return "", err
	}

	fmt.Printf("wrote %d bytes: %s\n", n, data)

	buf := make([]byte, 1024)
	n, addr, err := listenConn.ReadFromUDP(buf)
	if err != nil {
		return "", err
	}

	fmt.Printf("%d from %s: %s\n", n, addr, buf[0:n])

	if target != nil {
		err = xml.Unmarshal(buf[0:n], target)
		if err != nil {
			return "", err
		}
	}

	conn.Close()
	listenConn.Close()

	return string(buf[0:n]), nil
}

type EmotivaTransponder struct {
	XMLName      xml.Name                  `xml:"emotivaTransponder"`
	Model        string                    `xml:"model"`
	DataRevision string                    `xml:"dataRevision"`
	Name         string                    `xml:"name"`
	Control      EmotivaTransponderControl `xml:"control"`
}

type EmotivaTransponderControl struct {
	Version         string `xml:"version"`
	ControlPort     int    `xml:"controlPort"`
	NotifyPort      int    `xml:"notifyPort"`
	InfoPort        int    `xml:"infoPort"`
	SetupPortTCP    int    `xml:"setupPortTCP"`
	MenuNotifyPort  int    `xml:"menuNotifyPort"`
	SetupXMLVersion int    `xml:"setupXMLVersion"`
}

func emotivaPing(c *cli.Context) error {
	var response EmotivaTransponder

	_, err := emotivaRW(c, "<emotivaPing/>", "7000", "7001", &response)
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", response)

	return nil
}

func emotivaControl(c *cli.Context, command string) error {
	_, err := emotivaRW(c, fmt.Sprintf("<emotivaControl>%s</emotivaControl>", command), "7002", "7002", nil)
	if err != nil {
		return err
	}
	return nil
}

func inputFuncGenerator(input string) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		return emotivaControl(c, fmt.Sprintf(`<%s value="0" ack="yes" />`, input))
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
	err := emotivaControl(c, `<menu value="0" ack="yes" />`)
	if err != nil {
		return err
	}

	cc := make(chan os.Signal)
	signal.Notify(cc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cc
		err = emotivaControl(c, `<menu value="0" ack="yes" />`)
		if err != nil {
			panic(err)
		}
		exec.Command("stty", "-f", "/dev/tty", "echo").Run()
		os.Exit(1)
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
		fmt.Printf("b is %d\n", b)
		switch int(b) {
		case 65:
			// up
			err = emotivaControl(c, `<up value="0" ack="yes" />`)
		case 66:
			// down
			emotivaControl(c, `<down value="0" ack="yes" />`)
		case 67:
			// right
			emotivaControl(c, `<right value="0" ack="yes" />`)
		case 68:
			// left
			emotivaControl(c, `<left value="0" ack="yes" />`)
		case 10:
			// enter
			emotivaControl(c, `<enter value="0" ack="yes" />`)
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
				return emotivaControl(c, fmt.Sprintf(`<set_volume value="%d" ack="yes" />`, c.Int("volume")))
			},
		},
		{
			Name:  "menu",
			Usage: "toggle menu",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return emotivaControl(c, `<menu value="0" ack="yes" />`)
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
				return emotivaControl(c, `<info value="0" ack="yes" />`)
			},
		},
		{
			Name:  "ping",
			Usage: "emotivaPing",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return emotivaPing(c)
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
						return emotivaControl(c, `<power_on value="0" ack="yes" />`)
					},
				},
				{
					Name:  "off",
					Usage: "turn power off",
					Action: func(c *cli.Context) error {
						return emotivaControl(c, `<power_off value="0" ack="yes" />`)
					},
				},
				{
					Name:  "standby",
					Usage: "set standby",
					Action: func(c *cli.Context) error {
						return emotivaControl(c, `<standby value="0" ack="yes" />`)
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
						return emotivaControl(c, `<power_on value="0" ack="yes" />`)
					},
				},
				{
					Name:  "off",
					Usage: "turn power off",
					Action: func(c *cli.Context) error {
						return emotivaControl(c, `<power_off value="0" ack="yes" />`)
					},
				},
				{
					Name:  "standby",
					Usage: "set standby",
					Action: func(c *cli.Context) error {
						return emotivaControl(c, `<standby value="0" ack="yes" />`)
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
						return emotivaControl(c, `<loudness_on value="0" ack="yes" />`)
					},
				},
				{
					Name:  "off",
					Usage: "turn loudness off",
					Action: func(c *cli.Context) error {
						return emotivaControl(c, `<loudness_off value="0" ack="yes" />`)
					},
				},
				{
					Name:  "toggle",
					Usage: "toggle loudness",
					Action: func(c *cli.Context) error {
						return emotivaControl(c, `<loudness value="0" ack="yes" />`)
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

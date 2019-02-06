package main

import (
	"fmt"
	"net"
	"os"

	"github.com/urfave/cli"
)

func emotivaControl(c *cli.Context, command string) error {
	listenAddr, err := net.ResolveUDPAddr("udp", ":7002")
	if err != nil {
		return err
	}

	doneChan := make(chan bool)

	listenConn, err := net.ListenUDP("udp", listenAddr)
	go func(listenConn *net.UDPConn, doneChan chan<- bool) {
		buf := make([]byte, 1024)
		n, addr, err := listenConn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d from %s: %s\n", n, addr, buf[0:n])
		doneChan <- true
	}(listenConn, doneChan)

	conn, err := net.Dial("udp", "10.0.8.41:7002")
	if err != nil {
		return err
	}

	data := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?><emotivaControl>%s</emotivaControl>\n", command)

	n, err := conn.Write([]byte(data))
	if err != nil {
		return err
	}

	fmt.Printf("wrote %d bytes: %s\n", n, data)

	<-doneChan
	return nil
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Usage: "Emotiva address",
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
			Name:  "info",
			Usage: "display info screen",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return emotivaControl(c, `<info value="0" ack="yes" />`)
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

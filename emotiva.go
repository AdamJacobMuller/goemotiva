package emotiva

import (
	"encoding/xml"
	"fmt"
	"net"
)

type EmotivaController struct {
	address string

	controlTX      *net.UDPConn
	controlRX      *net.UDPConn
	pingTX         *net.UDPConn
	pingRX         *net.UDPConn
	NotifyPort     *net.Conn
	InfoPort       *net.Conn
	SetupPortTCP   *net.Conn
	MenuNotifyPort *net.Conn
}

func connPair(address string, tx string, rx string) (*net.UDPConn, *net.UDPConn, error) {
	rxAddr, err := net.ResolveUDPAddr("udp", ":"+rx)
	if err != nil {
		return nil, nil, err
	}

	rxConn, err := net.ListenUDP("udp", rxAddr)
	if err != nil {
		return nil, nil, err
	}

	txAddr, err := net.ResolveUDPAddr("udp", address+":"+tx)
	if err != nil {
		return nil, nil, err
	}

	txConn, err := net.DialUDP("udp", nil, txAddr)
	if err != nil {
		return nil, nil, err
	}

	return txConn, rxConn, nil
}

func NewEmotivaController(address string) (*EmotivaController, error) {
	ec := &EmotivaController{
		address: address,
	}

	pingTX, pingRX, err := connPair(ec.address, "7000", "7001")
	if err != nil {
		return nil, err
	}
	ec.pingTX = pingTX
	ec.pingRX = pingRX

	info, err := ec.Ping()
	if err != nil {
		return nil, err
	}

	controlTX, controlRX, err := connPair(ec.address, info.Control.ControlPort, info.Control.ControlPort)
	if err != nil {
		return nil, err
	}
	ec.controlTX = controlTX
	ec.controlRX = controlRX

	return ec, nil
}

func (ec *EmotivaController) Close() error {
	ec.controlTX.Close()
	ec.controlRX.Close()
	ec.pingTX.Close()
	ec.pingRX.Close()

	return nil
}

func (ec *EmotivaController) rw(body string, tx, rx *net.UDPConn, target interface{}) (string, error) {
	data := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>%s", body)

	fmt.Printf("writing %d bytes: %s\n", len(data), data)

	n, err := tx.Write([]byte(data))
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024)
	n, addr, err := rx.ReadFromUDP(buf)
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

	return string(buf[0:n]), nil
}

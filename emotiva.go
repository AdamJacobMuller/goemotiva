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
	notifyRX       *net.UDPConn
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

	notifyAddr, err := net.ResolveUDPAddr("udp", ":"+info.Control.NotifyPort)
	if err != nil {
		return nil, err
	}

	ec.notifyRX, err = net.ListenUDP("udp", notifyAddr)
	if err != nil {
		return nil, err
	}

	return ec, nil
}

func (ec *EmotivaController) Close() error {
	ec.controlTX.Close()
	ec.controlRX.Close()
	ec.pingTX.Close()
	ec.pingRX.Close()

	return nil
}

type Value struct {
	Value   string `xml:"value,attr"`
	Visible string `xml:"visible,attr"`
}
type Notify struct {
	XMLName            xml.Name `xml:"emotivaNotify"`
	Mode               Value    `xml:"mode"`
	SelectedMode       Value    `xml:"selected_mode"`
	Center             Value    `xml:"center"`
	Subwoofer          Value    `xml:"subwoofer"`
	Surround           Value    `xml:"surround"`
	Back               Value    `xml:"back"`
	Width              Value    `xml:"width"`
	Height             Value    `xml:"height"`
	ModeAuto           Value    `xml:"mode_auto"`
	ModeMusic          Value    `xml:"mode_music"`
	ModeMovie          Value    `xml:"mode_movie"`
	SelectedMovieMusic Value    `xml:"selected_movie_music"`
	ModeDirect         Value    `xml:"mode_direct"`
	ModeSurround       Value    `xml:"mode_surround"`
	ModeRefStereo      Value    `xml:"mode_ref_stereo"`
	ModeDolby          Value    `xml:"mode_dolby"`
	ModeDts            Value    `xml:"mode_dts"`
	ModeAllStereo      Value    `xml:"mode_all_stereo"`
	ModeStereo         Value    `xml:"mode_stereo"`
}

func (ec *EmotivaController) ReadNotify() (*Notify, error) {
	buf := make([]byte, 1024)
	fmt.Printf("ReadNotify\n")
	n, addr, err := ec.notifyRX.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}

	fmt.Printf("ReadNotify %d from %s at %s: %s\n", n, addr, ec.notifyRX.LocalAddr().String(), buf[0:n])

	np := &Notify{}

	err = xml.Unmarshal(buf[0:n], np)
	if err != nil {
		return nil, err
	}

	return np, nil
}

func (ec *EmotivaController) ReadControl() error {
	buf := make([]byte, 1024)
	fmt.Printf("ReadControl\n")
	n, addr, err := ec.controlRX.ReadFromUDP(buf)
	if err != nil {
		return err
	}

	fmt.Printf("ReadControl %d from %s at %s: %s\n", n, addr, ec.controlRX.LocalAddr().String(), buf[0:n])

	return nil
}

func (ec *EmotivaController) WriteControl(body string) error {
	data := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>%s", body)

	fmt.Printf("writing %d bytes: %s to %s\n", len(data), data, ec.controlTX.RemoteAddr().String())

	_, err := ec.controlTX.Write([]byte(data))
	if err != nil {
		return err
	}

	return nil
}

func (ec *EmotivaController) w(body string, tx *net.UDPConn) error {
	data := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>%s", body)

	fmt.Printf("writing %d bytes to %s: %s\n", len(data), tx.RemoteAddr().String(), data)

	_, err := tx.Write([]byte(data))
	if err != nil {
		return err
	}

	return nil
}

func (ec *EmotivaController) r(rx *net.UDPConn, target interface{}) (string, error) {
	buf := make([]byte, 1024)
	n, addr, err := rx.ReadFromUDP(buf)
	if err != nil {
		return "", err
	}

	fmt.Printf("read %d from %s at %s: %s\n", n, addr, rx.LocalAddr().String(), buf[0:n])

	if target != nil {
		err = xml.Unmarshal(buf[0:n], target)
		if err != nil {
			return "", err
		}
	}

	return string(buf[0:n]), nil
}

func (ec *EmotivaController) rw(body string, tx, rx *net.UDPConn, target interface{}) (string, error) {
	data := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>%s", body)

	fmt.Printf("writing %d bytes to %s: %s\n", len(data), tx.RemoteAddr().String(), data)

	n, err := tx.Write([]byte(data))
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024)
	n, addr, err := rx.ReadFromUDP(buf)
	if err != nil {
		return "", err
	}

	fmt.Printf("read %d from %s at %s: %s\n", n, addr, rx.LocalAddr().String(), buf[0:n])

	if target != nil {
		err = xml.Unmarshal(buf[0:n], target)
		if err != nil {
			return "", err
		}
	}

	return string(buf[0:n]), nil
}

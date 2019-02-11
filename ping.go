package emotiva

import (
	"encoding/xml"
)

func (ec *EmotivaController) Ping() (*EmotivaTransponder, error) {
	pr := &EmotivaTransponder{}

	_, err := ec.rw("<emotivaPing/>", ec.pingTX, ec.pingRX, pr)
	if err != nil {
		return nil, err
	}
	return pr, nil
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
	ControlPort     string `xml:"controlPort"`
	NotifyPort      string `xml:"notifyPort"`
	InfoPort        string `xml:"infoPort"`
	SetupPortTCP    string `xml:"setupPortTCP"`
	MenuNotifyPort  string `xml:"menuNotifyPort"`
	SetupXMLVersion string `xml:"setupXMLVersion"`
}

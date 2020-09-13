package emotiva

import (
	"fmt"
)

func (ec *EmotivaController) Status(commands []string, target interface{}) (string, error) {
	var command string
	for _, c := range commands {
		command = fmt.Sprintf("%s<%s/>", command, c)
	}
	body, err := ec.rw(fmt.Sprintf("<emotivaSubscription>%s</emotivaSubscription>", command), ec.controlTX, ec.controlRX, target)
	if err != nil {
		return "", err
	}

	return body, nil
}

func (ec *EmotivaController) Subscribe(commands []string, target interface{}) (string, error) {
	var command string
	for _, c := range commands {
		command = fmt.Sprintf("%s<%s/>", command, c)
	}
	body, err := ec.rw(fmt.Sprintf("<emotivaSubscription>%s</emotivaSubscription>", command), ec.controlTX, ec.controlRX, target)
	if err != nil {
		return "", err
	}

	return body, nil
}

func (ec *EmotivaController) Control(command string, target interface{}) (string, error) {
	body, err := ec.rw(fmt.Sprintf("<emotivaControl>%s</emotivaControl>", command), ec.controlTX, ec.controlRX, target)
	if err != nil {
		return "", err
	}

	return body, nil
}

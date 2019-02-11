package emotiva

import (
	"fmt"
)

func (ec *EmotivaController) Control(command string, target interface{}) (string, error) {
	body, err := ec.rw(fmt.Sprintf("<emotivaControl>%s</emotivaControl>", command), ec.controlTX, ec.controlRX, target)
	if err != nil {
		return "", err
	}

	return body, nil
}

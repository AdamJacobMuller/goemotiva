package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/gocraft/web"
	"github.com/satori/go.uuid"
)

var isUUID = regexp.MustCompile("^([0-9a-f]{8})-([0-9a-f]{4})-([0-9a-f]{4})-([0-9a-f]{4})-([0-9a-f]{12})$")

func newUUID() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

func unmarshal_json_request(rw web.ResponseWriter, req *web.Request, models ...interface{}) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	for _, model := range models {
		err = json.Unmarshal(data, &model)
		if err != nil {
			return err
		}
	}

	return nil
}

type RedactableModel interface {
	Redact() interface{}
}

func (c *Context) marshal_json_reply_unredacted(model RedactableModel) error {
	bytes, err := json.Marshal(model)
	if err != nil {
		return err
	}

	c.rw.Header().Set("content-type", "application/json")
	c.rw.WriteHeader(200)
	c.rw.Write(bytes)

	return nil
}

func (c *Context) marshal_json_reply(model RedactableModel) error {
	send := model.Redact()

	bytes, err := json.Marshal(send)
	if err != nil {
		return err
	}

	c.rw.Header().Set("content-type", "application/json")
	c.rw.WriteHeader(200)
	c.rw.Write(bytes)

	return nil
}

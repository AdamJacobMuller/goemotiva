package server

import (
	"fmt"
	"net/http"

	"github.com/adamjacobmuller/goemotiva"
	"github.com/adamjacobmuller/weblogrus"
	"github.com/gocraft/web"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	ec *emotiva.EmotivaController
}
type Context struct {
	server        *Server
	HTTPRequestID string
	log           *log.Entry
	rw            web.ResponseWriter
	req           *web.Request
}

func NewServer(address string) (*Server, error) {
	server := &Server{}

	ec, err := emotiva.NewEmotivaController(address)
	if err != nil {
		return nil, err
	}

	ec.WriteControl("<emotivaSubscription><video_format/><mode_ref_stereo/></emotivaSubscription>")
	go func() {
		for {
			n, err := ec.ReadNotify()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("ReadNotify failed")
				continue
			}
			fmt.Printf("NOTIFY: %#v\n", n)
		}
	}()

	go func() {
		for {
			ec.ReadControl()
		}
	}()

	publicRouter := web.New(Context{})
	publicRouter.Middleware(weblogrus.NewMiddleware().ServeHTTP)

	publicRouter.Middleware(func(c *Context, rw web.ResponseWriter,
		req *web.Request, next web.NextMiddlewareFunc) {
		c.server = server
		c.req = req
		c.rw = rw
		c.HTTPRequestID = newUUID()

		c.log = log.WithFields(log.Fields{
			"url":             req.URL,
			"method":          req.Method,
			"remote":          req.RemoteAddr,
			"http-request-id": c.HTTPRequestID,
		})

		next(rw, req)
	})

	httpServer := &http.Server{
		Addr:    ":8090",
		Handler: publicRouter,
	}
	err = httpServer.ListenAndServe()
	if err != nil {
		return nil, err
	}

	return server, nil
}

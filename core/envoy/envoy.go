package envoy

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Receiver represents a function that handles an incoming envoy request.
type Receiver func(data interface{}, res Responder)

// Responder represents a function that can respond to a specific envoy request.
type Responder func(res Response) error

// Request represents a request sent using envoy. It includes the service and route code which is
// used by envoy to determine how to route the request.
type Request struct {
	Service    int
	Route      int
	Data       interface{}
	ExpirySecs int
}

// requestData represents the data that is sent as part of a request. It contains the request data
// as well as the request id which is used to send back the response to the original entity that
// made the request.
type requestData struct {
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}

// Response represents an envoy response.
type Response struct {
	Data       interface{}
	ExpirySecs int
}

const (
	defaultRequestExpirySecs  = 1 * 60 // 1 minute
	defaultResponseExpirySecs = 1 * 60 // 1 minute
)

// Config represents configuration options for envoy.
type Config struct {
	Redis              string
	Service            int
	RequestExpirySecs  int
	ResponseExpirySecs int
}

// Envoy is a service that can handle server to server communication using redis as a message
// broker. Envoy handles routing requests and responses to the appropriate entities. For routing,
// requests, envoy uses a unique service and route code to determine how to dispatch the request.
// For routing responses, envoy uses a unique id that is attached to each request to determine
// where to send the response.
type Envoy struct {
	redis              *redis.Pool
	service            int
	requestExpirySecs  int
	responseExpirySecs int
}

// New creates a new envoy.
func New(cfg Config) (*Envoy, error) {
	r := &redis.Pool{
		IdleTimeout: 3 * time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(cfg.Redis)
		},
	}

	conn := r.Get()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		return nil, err
	}

	if cfg.RequestExpirySecs == 0 {
		cfg.RequestExpirySecs = defaultRequestExpirySecs
	}
	if cfg.ResponseExpirySecs == 0 {
		cfg.ResponseExpirySecs = defaultResponseExpirySecs
	}

	e := &Envoy{
		redis:              r,
		service:            cfg.Service,
		requestExpirySecs:  cfg.RequestExpirySecs,
		responseExpirySecs: cfg.ResponseExpirySecs,
	}
	return e, nil
}

// Send sends an envoy request.
func (e *Envoy) Send(req Request) (string, error) {
	conn := e.redis.Get()
	defer conn.Close()

	id := uuid.New().String()

	bytes, err := json.Marshal(requestData{id, req.Data})
	if err != nil {
		return "", err
	}

	if req.ExpirySecs == 0 {
		req.ExpirySecs = e.requestExpirySecs
	}

	recKey := fmt.Sprintf("%d:%d", req.Service, req.Route)

	if _, err := conn.Do("MULTI"); err != nil {
		return "", err
	}
	if _, err := conn.Do("RPUSH", recKey, bytes); err != nil {
		return "", err
	}
	if _, err := conn.Do("EXPIRE", recKey, req.ExpirySecs); err != nil {
		return "", err
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return "", err
	}
	return id, nil
}

// Receive designates a receiver to handle a particular route.
func (e *Envoy) Receive(route int, rec Receiver) error {
	conn := e.redis.Get()
	defer conn.Close()

	for {
		reply, err := redis.ByteSlices(conn.Do("BLPOP", fmt.Sprintf("%d:%d", e.service, route), 0))
		if err != nil {
			return err
		}

		var req requestData
		if err := json.Unmarshal(reply[1], &req); err != nil {
			return err
		}

		go rec(req.Data, e.responder(req))
	}
}

// Wait waits for a response for the request identified by the provided id. The provided timeout
// determines how long to wait before returning a timeout error. If the provided timeout is 0, then
// envoy waits indefinitely for a response.
func (e *Envoy) Wait(id string, timeoutSecs int) (string, error) {
	conn := e.redis.Get()
	defer conn.Close()

	reply, err := redis.Strings(conn.Do("BLPOP", id, timeoutSecs))
	if err != nil {
		return "", err
	}
	return reply[1], nil
}

func (e *Envoy) responder(req requestData) Responder {
	return func(res Response) error {
		conn := e.redis.Get()
		defer conn.Close()

		if res.ExpirySecs == 0 {
			res.ExpirySecs = e.responseExpirySecs
		}

		if _, err := conn.Do("MULTI"); err != nil {
			return err
		}
		if _, err := conn.Do("RPUSH", req.ID, res.Data); err != nil {
			return err
		}
		if _, err := conn.Do("EXPIRE", req.ID, res.ExpirySecs); err != nil {
			return err
		}
		if _, err := conn.Do("EXEC"); err != nil {
			return err
		}
		return nil
	}
}

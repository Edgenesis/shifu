package lwm2m

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/pion/dtls/v3"
	dtlsServer "github.com/plgd-dev/go-coap/v3/dtls/server"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	udpClient "github.com/plgd-dev/go-coap/v3/udp/client"
	udpServer "github.com/plgd-dev/go-coap/v3/udp/server"
)

type Server struct {
	router *mux.Router

	settings             v1alpha1.LwM2MSetting
	Conn                 mux.Conn
	endpointName         string
	lifeTime             int
	lastRegistrationTime time.Time
	deviceTokenMap       map[string]string // map[token]device
	observeCallback      map[string]func(interface{})
	onRegister           []func() error
}

const ()

const (
	LwM2MServerRegisterURI        = "/rd"
	LwM2MServerHandlerURI         = "/rd/{deviceId}"
	deviceId               string = "shifu"
	registerResponseValue  string = "rd"

	defaultLwM2MNetwork    string = "udp"
	defaultCoapListenPort  string = ":5683"
	defaultCoapsListenPort string = ":5684"
	keepAliveTimeoutSec           = 10 * 60
	keepAliveRetryTimes           = 10
)

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		logger.Debugf("ClientAddress %v, %v\n", w.Conn().RemoteAddr(), r.String())
		next.ServeCOAP(w, r)
	})
}

func (s *Server) Execute(objectId string, args string) error {
	req, err := s.Conn.NewPostRequest(s.Conn.Context(), objectId, message.TextPlain, strings.NewReader(args))
	if err != nil {
		return err
	}

	resp, err := s.Conn.Do(req)
	if err != nil {
		return err
	}

	if resp.Code() != codes.Changed {
		return errors.New("failed to execute object")
	}

	return nil
}

func NewServer(settings v1alpha1.LwM2MSetting) (*Server, error) {
	var server = &Server{
		endpointName:    settings.EndpointName,
		observeCallback: make(map[string]func(interface{})),
		deviceTokenMap:  make(map[string]string),
		settings:        settings,
	}

	router := mux.NewRouter()
	if err := errors.Join(
		router.Handle(LwM2MServerRegisterURI, mux.HandlerFunc(server.handleRegister)),
		router.Handle(LwM2MServerHandlerURI, mux.HandlerFunc(server.handleResourceUpdate)),
	); err != nil {
		return nil, err
	}

	router.DefaultHandle(mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		token := r.Token()
		deviceId, exists := server.deviceTokenMap[string(token)]
		if exists {
			if fn, exists := server.observeCallback[deviceId]; exists {
				data, err := io.ReadAll(r.Body())
				if err != nil {
					_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("failed to read body")))
					return
				}
				fn(string(data))
			}
		}
	}))

	router.Use(loggingMiddleware)
	server.router = router
	return server, nil
}

func (s *Server) Run() error {
	switch *s.settings.SecurityMode {
	case v1alpha1.SecurityModeDTLS:
		return s.startDTLSServer()
	default:
		logger.Infof("securityMode not set or not support, using none security mode")
	// default using none security mode
	case v1alpha1.SecurityModeNone:
	}
	return s.startUDPServer()
}

func (s *Server) startDTLSServer() error {
	switch *s.settings.DTLSMode {
	case v1alpha1.DTLSModePSK:
		serverOptions := []dtlsServer.Option{
			options.WithMux(s.router),
			options.WithContext(context.Background()),
			options.WithKeepAlive(keepAliveRetryTimes, time.Second*keepAliveTimeoutSec, func(cc *udpClient.Conn) {
				// TODO: handle inactive connection
			}),
		}

		server := dtlsServer.New(serverOptions...)

		cipherSuites, err := CipherSuiteStringsToCodes(s.settings.CipherSuites)
		if err != nil {
			return err
		}

		psk, err := hex.DecodeString(*s.settings.PSKKey)
		if err != nil {
			return err
		}

		dtlsConfig := dtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				return []byte(psk), nil
			},
			PSKIdentityHint: []byte(*s.settings.PSKIdentity),
			CipherSuites:    cipherSuites,
		}

		l, err := net.NewDTLSListener(defaultLwM2MNetwork, defaultCoapsListenPort, &dtlsConfig)
		if err != nil {
			return err
		}

		return server.Serve(l)
	case v1alpha1.DTLSModeRPK:
		fallthrough
	case v1alpha1.DTLSModeX509:
		return errors.New("not implemented")
	default:
		logger.Infof("dtlsMode not set, using none security mode")
		// default using none security mode
	}

	return s.startUDPServer()
}

func (s *Server) startUDPServer() error {
	serverOptions := []udpServer.Option{
		options.WithMux(s.router),
		options.WithContext(context.Background()),
		options.WithKeepAlive(keepAliveRetryTimes, time.Second*keepAliveRetryTimes, func(cc *udpClient.Conn) {
			logger.Error("inactive connection")
		}),
	}

	server := udpServer.New(serverOptions...)
	conn, err := net.NewListenUDP(defaultLwM2MNetwork, defaultCoapListenPort)
	if err != nil {
		return err
	}
	return server.Serve(conn)
}

func (s *Server) handleRegister(w mux.ResponseWriter, r *mux.Message) {
	// TODO: parse register message to get object links
	query, err := r.Queries()
	if err != nil {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("failed to read queries")))
		return
	}

	parsedQuery, err := parseRegisterQuery(query)
	if err != nil {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("failed to parse queries")))
		return
	}

	if parsedQuery.EndpointName != s.endpointName {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("endpoint name mismatch")))
		return
	}

	s.lifeTime, _ = strconv.Atoi(parsedQuery.Lifetime)
	if err := w.SetResponse(codes.Created, message.TextPlain, nil,
		message.Option{ID: message.LocationPath, Value: []byte(registerResponseValue)},
		message.Option{ID: message.LocationPath, Value: []byte(deviceId)},
	); err != nil {
		logger.Debug("register response failed")
	}

	s.lastRegistrationTime = time.Now()
	if s.Conn != nil {
		// try to close the previous connection if exists
		if err := s.Conn.Close(); err != nil {
			// log error but continue
			logger.Errorf("failed to close connection, error: %v", err)
		}
	}
	s.Conn = w.Conn()

	for _, fn := range s.onRegister {
		if err := fn(); err != nil {
			logger.Errorf("failed when calling register callback, error: %s", err.Error())
			_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("failed to register object links")))
			return
		}
	}
}

func (s *Server) OnRegister(fn func() error) {
	s.onRegister = append(s.onRegister, fn)
}

const (
	deviceIdObjectParam = "deviceId"
)

// handleResourceUpdate handles UPDATE and De-register request
func (s *Server) handleResourceUpdate(w mux.ResponseWriter, r *mux.Message) {
	deviceIdQuery := r.RouteParams.Vars[deviceIdObjectParam]
	if deviceIdQuery != deviceId {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("device id mismatch")))
		return
	}

	switch r.Code() {
	// De-register
	case codes.DELETE:
		if err := s.Conn.Close(); err != nil {
			logger.Errorf("failed to close connection, error: %v", err)
		}
		s.Conn = nil
		s.lastRegistrationTime = time.Time{}
		return
	// Update
	case codes.POST:
		// if not registered, handle register
		if s.Conn == nil {
			_ = w.SetResponse(codes.NotFound, message.TextPlain, nil)
			return
		}
		s.lastRegistrationTime = time.Now()
		// check if the request is from the same connection
		if s.Conn.RemoteAddr() != w.Conn().RemoteAddr() {
			_ = w.SetResponse(codes.BadRequest, message.TextPlain, nil)
			return
		}
		_ = w.SetResponse(codes.Changed, message.TextPlain, nil)
	default:
	}

}

type RegisterQuery struct {
	EndpointName string `json:"ep"`
	LwM2MVersion string `json:"lwm2m"`
	Lifetime     string `json:"lt"`
	BindingMode  string `json:"bnd"`
	SMSNumber    string `json:"sms"`
	ObjectLinks  string `json:"b"`
}

// parseRegisterQuery parses the register query string into a RegisterQuery struct
// example: "ep=test&lwm2m=1.0.3&lt=86400&bnd=U&sms=1234567890&b=1"
func parseRegisterQuery(queries []string) (*RegisterQuery, error) {
	var queryMap = make(map[string]string)
	for _, query := range queries {
		kvPair := strings.Split(query, "=")
		if len(kvPair) != 2 {
			// skip invalid query
			continue
		}
		queryMap[kvPair[0]] = kvPair[1]
	}

	// convert map to json
	data, err := json.Marshal(queryMap)
	if err != nil {
		return nil, err
	}

	var registerQuery RegisterQuery
	if err := json.Unmarshal(data, &registerQuery); err != nil {
		return nil, err
	}

	return &registerQuery, nil
}

// Read reads the object value from the server
// objectId: the object id to read example: "/3442/0/120"
func (s *Server) Read(objectId string) (string, error) {
	if err := s.checkRegistrationStatus(); err != nil {
		return "", err
	}

	request, err := s.Conn.NewGetRequest(s.Conn.Context(), objectId)
	if err != nil {
		return "", err
	}
	request.SetAccept(message.TextPlain)

	resp, err := s.Conn.Do(request)
	if err != nil {
		return "", err
	}

	if resp.Code() == codes.NotFound {
		return "", errors.New("object not found")
	}

	data, err := io.ReadAll(resp.Body())
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Write writes the object value to the server
// objectId: the object id to write example: "/3442/0/120"
// newValue: the new value to write
func (s *Server) Write(objectId string, newValue string) error {
	if err := s.checkRegistrationStatus(); err != nil {
		return err
	}

	request, err := s.Conn.NewPutRequest(s.Conn.Context(), objectId, message.TextPlain, strings.NewReader(newValue))
	if err != nil {
		return err
	}

	resp, err := s.Conn.Do(request)
	if err != nil {
		return err
	}

	if resp.Code() == codes.MethodNotAllowed {
		return errors.New("write method not allowed")
	}

	if resp.Code() != codes.Changed {
		return errors.New("failed to write object")
	}

	return nil
}

func (s *Server) Observe(objectId string, callback func(newData interface{})) error {
	if err := s.checkRegistrationStatus(); err != nil {
		return err
	}

	request, err := s.Conn.NewObserveRequest(s.Conn.Context(), objectId)
	if err != nil {
		return err
	}
	request.SetAccept(message.TextPlain)

	resp, err := s.Conn.Do(request)
	if err != nil {
		return err
	}

	if resp.Code() != codes.Content {
		return errors.New("failed to observe object")
	}
	token := resp.Token()
	s.observeCallback[objectId] = callback
	s.deviceTokenMap[string(token)] = objectId

	logger.Debugf("observe %s with token %s", objectId, token)
	return nil
}

func (s *Server) checkRegistrationStatus() error {
	if time.Since(s.lastRegistrationTime) > time.Second*time.Duration(s.lifeTime) {
		// TODO: handle re-register
		return errors.New("device is offline")
	}

	return nil
}

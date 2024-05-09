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
	"github.com/pion/dtls/v2"
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

	settings             v1alpha1.LwM2MSettings
	Conn                 mux.Conn
	endpointName         string
	liftTime             int
	lastRegistrationTime time.Time
	deviceTokenMap       map[string]string // map[token]device
	observeCallback      map[string]func(interface{})
	onRegister           []func() error
}

const deviceId string = "shifu"

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

func NewServer(settings v1alpha1.LwM2MSettings) (*Server, error) {
	var server = &Server{
		endpointName:    settings.EndpointName,
		observeCallback: make(map[string]func(interface{})),
		deviceTokenMap:  make(map[string]string),
		settings:        settings,
	}

	router := mux.NewRouter()
	if err := errors.Join(
		router.Handle("/rd", mux.HandlerFunc(server.handleRegister)),
		router.Handle("/rd/{deviceId}", mux.HandlerFunc(server.handleResource)),
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
			options.WithKeepAlive(10, time.Minute*10, func(cc *udpClient.Conn) {}),
		}

		server := dtlsServer.New(serverOptions...)

		cipersuites, err := StringsToCodes(s.settings.CipherSuites)
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
			CipherSuites:    cipersuites,
		}

		l, err := net.NewDTLSListener("udp4", ":5684", &dtlsConfig)
		if err != nil {
			return err
		}

		return server.Serve(l)
	case v1alpha1.DTLSModeRPK:
		fallthrough
	case v1alpha1.DTLSModeX509:
		return errors.New("not implemented")
	default:
		// default using none security mode
	}

	logger.Infof("dtlsMode not set, using none security mode")
	return s.startUDPServer()
}

func (s *Server) startUDPServer() error {
	serverOptions := []udpServer.Option{
		options.WithMux(s.router),
		options.WithContext(context.Background()),
		options.WithKeepAlive(10, time.Minute*10, func(cc *udpClient.Conn) {}),
	}

	server := udpServer.New(serverOptions...)
	conn, err := net.NewListenUDP("udp", ":5683")
	if err != nil {
		return err
	}
	return server.Serve(conn)
}

func (s *Server) handleRegister(w mux.ResponseWriter, r *mux.Message) {
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

	s.liftTime, _ = strconv.Atoi(parsedQuery.Lifetime)
	if err := w.SetResponse(codes.Created, message.TextPlain, nil,
		message.Option{ID: message.LocationPath, Value: []byte("rd")},
		message.Option{ID: message.LocationPath, Value: []byte(deviceId)},
	); err != nil {
		logger.Debug("register response failed")
	}

	s.lastRegistrationTime = time.Now()
	s.Conn = w.Conn()

	for _, fn := range s.onRegister {
		if err := fn(); err != nil {
			logger.Debug(err)
			_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("failed to register object links")))
			return
		}
	}
}

func (s *Server) OnRegister(fn func() error) {
	s.onRegister = append(s.onRegister, fn)
}

func (s *Server) handleResource(w mux.ResponseWriter, r *mux.Message) {
	deviceIdQuery := r.RouteParams.Vars["deviceId"]
	if deviceIdQuery != deviceId {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("device id mismatch")))
		return
	}

	switch r.Code() {
	case codes.DELETE:
		s.Conn = nil
		s.lastRegistrationTime = time.Time{}
		return
	case codes.POST:
		s.lastRegistrationTime = time.Now()
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

func DoNothing(newData interface{}) {}

func (s *Server) checkRegistrationStatus() error {
	if time.Since(s.lastRegistrationTime) > time.Second*time.Duration(s.liftTime) {
		return errors.New("device is offline")
	}

	return nil
}

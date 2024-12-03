package client

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/deviceshifu/deviceshifulwm2m/lwm2m"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	piondtls "github.com/pion/dtls/v3"
	"github.com/plgd-dev/go-coap/v3/dtls"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
	udpClient "github.com/plgd-dev/go-coap/v3/udp/client"
)

const (
	lwM2MVersion       = "1.0"
	defaultBindingMode = "U"

	rootObjectId = "root"
	registerPath = "/rd"

	observeTaskSuffix = "-ob"
)

type Client struct {
	ctx context.Context
	Config

	locationPath     string
	object           Object
	lastModifiedTime time.Time
	lastUpdatedTime  time.Time
	dataCache        map[string]interface{}

	udpConnection *udpClient.Conn
	taskManager   *TaskManager
}

type Config struct {
	EndpointName    string
	EndpointUrl     string
	DeviceShifuHost string
	Settings        v1alpha1.LwM2MSetting
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	var client = &Client{
		ctx:         context.TODO(),
		Config:      config,
		object:      *NewObject(rootObjectId, nil),
		taskManager: NewTaskManager(ctx),
		dataCache:   make(map[string]interface{}),
	}

	return client, nil
}

func (c *Client) Start() error {
	udpClientOpts := []udp.Option{}

	udpClientOpts = append(
		udpClientOpts,
		options.WithInactivityMonitor(time.Minute, func(cc *udpClient.Conn) {
			_ = cc.Close()
		}),
		options.WithMux(c.handleRouter()),
	)

	cipherSuites, err := lwm2m.CipherSuiteStringsToCodes(c.Settings.CipherSuites)
	if err != nil {
		return err
	}

	var conn *udpClient.Conn
	switch *c.Settings.SecurityMode {
	case v1alpha1.SecurityModeDTLS:
		switch *c.Settings.DTLSMode {
		case v1alpha1.DTLSModePSK:
			dtlsConfig := &piondtls.Config{
				PSK: func(hint []byte) ([]byte, error) {
					return hex.DecodeString(*c.Settings.PSKKey)
				},
				PSKIdentityHint: []byte(*c.Settings.PSKIdentity),
				CipherSuites:    cipherSuites,
			}

			conn, err = dtls.Dial(c.EndpointUrl, dtlsConfig, udpClientOpts...)
		}
	default:
		fallthrough
	case v1alpha1.SecurityModeNone:
		conn, err = udp.Dial(c.EndpointUrl, udpClientOpts...)
	}
	if err != nil {
		return err
	}

	c.udpConnection = conn
	return nil
}

func (c *Client) Object() Object {
	return c.object
}

type QueryParams string

const (
	QueryParamsEndpointName QueryParams = "ep"
	QueryParamslifeTime     QueryParams = "lt"
	QueryParamsLwM2MVersion QueryParams = "lwm2m"
	QueryParamsBindingMode  QueryParams = "b"
)

// Register register the client to the server
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=27
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=76
func (c *Client) Register() error {
	coRELinkStr := c.object.GetCoRELinkString()
	request, err := c.udpConnection.NewPostRequest(context.TODO(), registerPath, message.AppLinkFormat, strings.NewReader(coRELinkStr))
	if err != nil {
		return err
	}

	// set query params for register request
	// example: /rd?ep=shifu-gateway&lt=300&lwm2m=1.0&b=U
	request.AddQuery(fmt.Sprintf("%s=%s", QueryParamsEndpointName, c.EndpointName))
	request.AddQuery(fmt.Sprintf("%s=%d", QueryParamslifeTime, c.Settings.LifeTimeSec))
	request.AddQuery(fmt.Sprintf("%s=%s", QueryParamsLwM2MVersion, lwM2MVersion))
	request.AddQuery(fmt.Sprintf("%s=%s", QueryParamsBindingMode, defaultBindingMode))
	// only accept text/plain
	request.SetAccept(message.TextPlain)
	resp, err := c.udpConnection.Do(request)
	if err != nil {
		return err
	}

	if resp.Code() != codes.Created {
		return fmt.Errorf("register failed: %v", resp.Code())
	}

	locationPath, err := resp.Options().LocationPath()
	if err != nil {
		return err
	}

	c.locationPath = locationPath
	c.lastUpdatedTime = time.Now()
	go func() {
		if err := c.AutoUpdate(); err != nil {
			logger.Errorf("failed to auto update registration: %v", err)
		}
	}()

	logger.Infof("register %v success", c.locationPath)
	return nil
}

// De-register the client from the server
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=76
func (c *Client) Delete() error {
	request, err := c.udpConnection.NewDeleteRequest(context.Background(), c.locationPath)
	if err != nil {
		return err
	}

	resp, err := c.udpConnection.Do(request)
	if err != nil {
		return err
	}

	if resp.Code() != codes.Deleted {
		return errors.New("delete failed")
	}

	logger.Infof("delete %v success", c.locationPath)
	return nil
}

// AutoUpdate auto update registration
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=30
func (c *Client) AutoUpdate() error {
	ticker := time.NewTicker(time.Duration(c.Settings.UpdateIntervalSec) * time.Second)
	for {
		select {
		case <-c.ctx.Done():
			return nil
		case <-ticker.C:
			if c.isActivity() {
				if err := c.Update(); err != nil {
					logger.Errorf("failed to update registration: %v", err)
					continue
				}
				logger.Debug("update registration success")
			}
		}
	}
}

// Update update registration
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=30
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=76
func (c *Client) Update() error {
	var coRELinkStr string
	// If there are changes to the object, the CoRELinkStr should be updated in the payload
	if c.lastUpdatedTime.Before(c.lastModifiedTime) {
		coRELinkStr = c.object.GetCoRELinkString()
	} else {
		logger.Debug("update with no data changed")
	}

	resp, err := c.udpConnection.Post(c.ctx, c.locationPath, message.AppLinkFormat, strings.NewReader(coRELinkStr))
	if err != nil {
		return err
	}

	if resp.Code() != codes.Changed {
		return errors.New("update failed")
	}

	c.lastUpdatedTime = time.Now()
	return nil
}

func (c *Client) handleRouter() *mux.Router {
	router := mux.NewRouter()
	// default to handle object request like read, write and execute
	router.DefaultHandle(mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		if r.Type() == message.Reset {
			// ping response is reset message, ignore it
			return
		}

		objectId, err := r.Path()
		if err != nil {
			_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
		}

		// get object which is requested
		object := c.object.GetChildObject(objectId)
		if object == nil {
			_ = w.SetResponse(codes.NotFound, message.TextPlain, nil)
			return
		}

		switch r.Code() {
		case codes.GET:
			// read data from object
			// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=33
			// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=78
			// if observe option is set, then handle observe action
			if r.Options().HasOption(message.Observe) {
				c.handleObserve(w, r)
				return
			}

			res, err := c.object.ReadAll(objectId)
			if err != nil {
				logger.Errorf("failed to read data from object %s, error: %v", objectId, err)
				_ = w.SetResponse(codes.NotFound, message.TextPlain, strings.NewReader(err.Error()))
				return
			}
			_ = w.SetResponse(codes.Content, message.AppLwm2mJSON, strings.NewReader(res.ReadAsJSON()))
			return
		case codes.PUT:
			// write data to object
			// read data from request body and write to object
			// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=33
			// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=78
			newData, err := io.ReadAll(r.Body())
			if err != nil {
				_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
				return
			}
			err = object.Write(string(newData))
			if err != nil {
				_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
				return
			}
			_ = w.SetResponse(codes.Changed, message.TextPlain, nil)

		case codes.POST:
			// execute object
			// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=78
			err = object.Execute()
			if err != nil {
				_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
				return
			}

			_ = w.SetResponse(codes.Changed, message.TextPlain, nil)

		default:
			_ = w.SetResponse(codes.MethodNotAllowed, message.TextPlain, nil)
		}

	}))

	return router
}

func (c *Client) AddObject(object Object) {
	logger.Infof("add object %v", object.Id)
	// check if object already exists
	if obj, exists := c.object.Child[object.Id]; exists {
		// if object already exists, add object to target path
		obj.AddObject(object.Id, object)
	} else {
		// if object not exists, then add object to the root object
		c.object.AddGroup(object)
	}

	c.lastModifiedTime = time.Now()
}

func (c *Client) Ping() error {
	return c.udpConnection.Ping(c.ctx)
}

const (
	EnableObserveAction  uint32 = 0
	DisableObserveAction uint32 = 1
)

// handleObserve handle observe action
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=37
// Reference: https://www.openmobilealliance.org/release/LightweightM2M/V1_0-20170208-A/OMA-TS-LightweightM2M-V1_0-20170208-A.pdf#page=80
func (c *Client) handleObserve(w mux.ResponseWriter, r *mux.Message) {
	objectId, err := r.Path()
	if err != nil {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
		return
	}

	logger.Debugf("observe %v", objectId)

	observeAction, err := r.Options().GetUint32(message.Observe)
	if err != nil {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
		return
	}
	switch observeAction {
	case EnableObserveAction:
		c.observe(w, r.Token(), objectId)
	case DisableObserveAction:
		c.cancelObserve(w, objectId)
	}
}

func (c *Client) observe(w mux.ResponseWriter, token message.Token, objectId string) {
	// start obs with 2 seq number
	var obs uint32 = 2
	// report new data with interval 30s
	// TODO need to config it by read Attribute from object

	c.taskManager.AddTask(objectId, time.Second*time.Duration(c.Settings.ObserveIntervalSec), func() {
		data, err := c.object.ReadAll(objectId)
		if err != nil {
			return
		}

		jsonData := data.ReadAsJSON()

		c.dataCache[objectId] = jsonData
		err = sendResponse(w.Conn(), token, obs, jsonData)
		if err != nil {
			logger.Errorf("failed to send response: %v", err)
			return
		}
		obs++
		// reset data changed notify task to avoid data changed notify too frequently
		c.taskManager.ResetTask(objectId + observeTaskSuffix)
	})

	// report new data with a interval 5s to check data is changed
	c.taskManager.AddTask(objectId+observeTaskSuffix, time.Second*5, func() {
		data, err := c.object.ReadAll(objectId)
		if err != nil {
			logger.Errorf("failed to read data from object %s, error: %v", objectId, err)
			return
		}

		jsonData := data.ReadAsJSON()

		// check data is changed
		if data, exists := c.dataCache[objectId]; exists {
			if string(jsonData) == data {
				logger.Debug("no data changed")
				return
			}
		}

		c.dataCache[objectId] = jsonData
		err = sendResponse(w.Conn(), token, obs, jsonData)
		if err != nil {
			return
		}
		obs++
		c.taskManager.ResetTask(objectId)
	})

	res, err := c.object.ReadAll(objectId)
	if err != nil {
		_ = w.SetResponse(codes.NotFound, message.TextPlain, nil)
		return
	}

	jsonData := res.ReadAsJSON()
	c.dataCache[objectId] = string(jsonData)
	_ = w.SetResponse(codes.Content, message.AppLwm2mJSON, strings.NewReader(jsonData),
		message.Option{ID: message.Observe, Value: []byte{byte(obs)}},
	)
}

func (c *Client) cancelObserve(w mux.ResponseWriter, objectId string) {
	logger.Infof("cancel observe %v", objectId)
	c.taskManager.CancelTask(objectId)
	c.taskManager.CancelTask(objectId + observeTaskSuffix)
	res, err := c.object.ReadAll(objectId)
	if err != nil {
		logger.Errorf("failed to read data from object %s, error: %v", objectId, err)
		_ = w.SetResponse(codes.NotFound, message.TextPlain, strings.NewReader(err.Error()))
		return
	}
	_ = w.SetResponse(codes.Content, message.AppLwm2mJSON, strings.NewReader(res.ReadAsJSON()))
}

func sendResponse(cc mux.Conn, token []byte, obs uint32, body string) error {
	m := cc.AcquireMessage(cc.Context())
	defer cc.ReleaseMessage(m)
	m.SetCode(codes.Content)
	m.SetToken(token)
	m.SetBody(strings.NewReader(body))
	m.SetContentFormat(message.AppLwm2mJSON)
	m.SetObserve(obs)
	return cc.WriteMessage(m)
}

func (c *Client) CleanUp() {
	c.taskManager.CancelAllTasks()
	_ = c.Delete()
}

func (c *Client) isActivity() bool {
	return time.Now().Before(c.lastUpdatedTime.Add(time.Duration(c.Settings.LifeTimeSec) * time.Second))
}

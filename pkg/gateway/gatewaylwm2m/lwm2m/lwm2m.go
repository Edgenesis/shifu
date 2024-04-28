package lwm2m

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
	"github.com/plgd-dev/go-coap/v3/udp/client"
)

type Client struct {
	ctx context.Context

	endpointName string
	serverUrl    string
	locationPath string

	updateInterval   int
	liftTime         int
	object           Object
	lastModifiedTime time.Time
	lastUpdatedTime  time.Time
	dataCache        map[string]interface{}

	conn *client.Conn
	tmgr *TaskManager
}

const (
	DefaultLifeTime       = 300
	DefaultUpdateInterval = 60
)

func NewClient(serverUrl string, endpointName string) (*Client, error) {
	var client = &Client{
		ctx:            context.TODO(),
		serverUrl:      serverUrl,
		endpointName:   endpointName,
		liftTime:       DefaultLifeTime,
		updateInterval: DefaultUpdateInterval,
		object:         *NewObject("root", nil),
		tmgr:           NewTaskManager(),
		dataCache:      make(map[string]interface{}),
	}
	udpClientOpts := []udp.Option{}

	udpClientOpts = append(udpClientOpts,
		options.WithMux(client.handleRouter()),
	)

	co, err := udp.Dial(serverUrl, udpClientOpts...)
	if err != nil {
		return nil, err
	}

	client.conn = co
	return client, nil
}

func (c *Client) Object() Object {
	return c.object
}

func (c *Client) Register() error {
	coRELinkStr := c.object.GetCoRELinkString()
	request, err := c.conn.NewPostRequest(context.TODO(), "/rd", message.AppLinkFormat, strings.NewReader(coRELinkStr))
	if err != nil {
		return err
	}

	request.AddQuery("ep=" + c.endpointName)
	request.AddQuery(fmt.Sprintf("lt=%d", c.liftTime))
	request.AddQuery("lwm2m=1.0")
	request.AddQuery("b=U")
	request.SetAccept(message.TextPlain)
	resp, err := c.conn.Do(request)
	if err != nil {
		return err
	}

	if resp.Code() != codes.Created {
		return errors.New("register failed")
	}

	locationPath, err := resp.Options().LocationPath()
	if err != nil {
		return err
	}

	c.locationPath = locationPath
	c.lastUpdatedTime = time.Now()
	go func() {
		panic(c.AutoUpdate())
	}()
	logger.Infof("register %v success", c.locationPath)
	return nil
}

func (c *Client) Delete() error {
	request, err := c.conn.NewDeleteRequest(context.Background(), c.locationPath)
	if err != nil {
		return err
	}

	resp, err := c.conn.Do(request)
	if err != nil {
		return err
	}

	if resp.Code() != codes.Deleted {
		return errors.New("delete failed")
	}

	logger.Infof("delete %v success", c.locationPath)
	return nil
}

func (c *Client) AutoUpdate() error {
	ticker := time.NewTicker(time.Duration(c.updateInterval) * time.Second)
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

func (c *Client) Update() error {
	var coRELinkStr string
	// if have changed of the object should set the CoRELinkStr updated in payload
	if c.lastUpdatedTime.Before(c.lastModifiedTime) {
		coRELinkStr = c.object.GetCoRELinkString()
	} else {
		logger.Info("no data changed")
	}

	resp, err := c.conn.Post(context.TODO(), c.locationPath, message.AppLinkFormat, strings.NewReader(coRELinkStr))
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
	router.DefaultHandle(mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		if r.Type() == message.Reset {
			c.tmgr.CancelAllTasks()
			return
		}

		objectId, err := r.Path()
		if err != nil {
			_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
		}

		switch r.Code() {
		case codes.GET:
			if r.Options().HasOption(message.Observe) {
				c.handleObserve(w, r)
				return
			}

			res, err := c.object.ReadAll(objectId)
			if err != nil {
				_ = w.SetResponse(codes.NotFound, message.TextPlain, strings.NewReader(err.Error()))
				return
			}
			_ = w.SetResponse(codes.Content, message.AppLwm2mJSON, strings.NewReader(res.ReadAsJSON()))
			return
		case codes.PUT:
			newData, err := io.ReadAll(r.Body())
			if err != nil {
				_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
				return
			}
			err = c.object.Write(string(newData))
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
	c.object.AddGroup(object)
	c.lastModifiedTime = time.Now()
}

func (c *Client) handleObserve(w mux.ResponseWriter, r *mux.Message) {
	objectId, err := r.Path()
	if err != nil {
		_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
		return
	}

	logger.Debugf("observe %v", objectId)
	token := r.Token()
	var obs uint32 = 2
	c.tmgr.AddTask(objectId, time.Second*10, func() {
		data, err := c.object.ReadAll(objectId)
		if err != nil {
			return
		}

		jsonData := data.ReadAsJSON()

		c.dataCache[objectId] = string(jsonData)
		err = sendResponse(w.Conn(), token, obs, jsonData)
		if err != nil {
			return
		}
		obs++
		c.tmgr.ResetTask(objectId + "-ob")
	})

	c.tmgr.AddTask(objectId+"-ob", time.Second*5, func() {
		data, err := c.object.ReadAll(objectId)
		if err != nil {
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

		c.dataCache[objectId] = string(jsonData)
		err = sendResponse(w.Conn(), token, obs, jsonData)
		if err != nil {
			return
		}
		obs++
		c.tmgr.ResetTask(objectId)
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
	c.tmgr.CancelAllTasks()
	_ = c.Delete()
}

func (c *Client) isActivity() bool {
	return time.Now().Before(c.lastUpdatedTime.Add(time.Duration(c.liftTime) * time.Second))
}

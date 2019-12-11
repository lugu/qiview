package main

import (
	"bytes"
	"context"
	"fmt"
	bus "github.com/lugu/qiloop/bus"
	net "github.com/lugu/qiloop/bus/net"
	basic "github.com/lugu/qiloop/type/basic"
	object "github.com/lugu/qiloop/type/object"
	value "github.com/lugu/qiloop/type/value"
)

// ALVideoDeviceImplementor interface of the service implementation
type ALVideoDeviceImplementor interface {
	// Activate is called before any other method.
	// It shall be used to initialize the interface.
	// activation provides runtime informations.
	// activation.Terminate() unregisters the object.
	// activation.Session can access other services.
	// helper enables signals and properties updates.
	// Properties must be initialized using helper,
	// during the Activate call.
	Activate(activation bus.Activation, helper ALVideoDeviceSignalHelper) error
	OnTerminate()
	SubscribeCamera(name string, cameraIndex int32, resolution int32, colorSpace int32, fps int32) (string, error)
	GetImageRemote(name string) (value.Value, error)
	Unsubscribe(nameId string) (bool, error)
}

// ALVideoDeviceSignalHelper provided to ALVideoDevice a companion object
type ALVideoDeviceSignalHelper interface{}

// stubALVideoDevice implements server.Actor.
type stubALVideoDevice struct {
	impl      ALVideoDeviceImplementor
	session   bus.Session
	service   bus.Service
	serviceID uint32
	signal    bus.SignalHandler
}

// ALVideoDeviceObject returns an object using ALVideoDeviceImplementor
func ALVideoDeviceObject(impl ALVideoDeviceImplementor) bus.Actor {
	var stb stubALVideoDevice
	stb.impl = impl
	obj := bus.NewBasicObject(&stb, stb.metaObject(), stb.onPropertyChange)
	stb.signal = obj
	return obj
}

// CreateALVideoDevice registers a new object to a service
// and returns a proxy to the newly created object
func CreateALVideoDevice(session bus.Session, service bus.Service, impl ALVideoDeviceImplementor) (ALVideoDeviceProxy, error) {
	obj := ALVideoDeviceObject(impl)
	objectID, err := service.Add(obj)
	if err != nil {
		return nil, err
	}
	stb := &stubALVideoDevice{}
	meta := object.FullMetaObject(stb.metaObject())
	client := bus.DirectClient(obj)
	proxy := bus.NewProxy(client, meta, service.ServiceID(), objectID)
	return MakeALVideoDevice(session, proxy), nil
}
func (p *stubALVideoDevice) Activate(activation bus.Activation) error {
	p.session = activation.Session
	p.service = activation.Service
	p.serviceID = activation.ServiceID
	return p.impl.Activate(activation, p)
}
func (p *stubALVideoDevice) OnTerminate() {
	p.impl.OnTerminate()
}
func (p *stubALVideoDevice) Receive(msg *net.Message, from bus.Channel) error {
	// action dispatch
	switch msg.Header.Action {
	case 100:
		return p.SubscribeCamera(msg, from)
	case 101:
		return p.GetImageRemote(msg, from)
	case 116:
		return p.Unsubscribe(msg, from)
	default:
		return from.SendError(msg, bus.ErrActionNotFound)
	}
}
func (p *stubALVideoDevice) onPropertyChange(name string, data []byte) error {
	switch name {
	default:
		return fmt.Errorf("unknown property %s", name)
	}
}
func (p *stubALVideoDevice) SubscribeCamera(msg *net.Message, c bus.Channel) error {
	buf := bytes.NewBuffer(msg.Payload)
	name, err := basic.ReadString(buf)
	if err != nil {
		return c.SendError(msg, fmt.Errorf("cannot read name: %s", err))
	}
	cameraIndex, err := basic.ReadInt32(buf)
	if err != nil {
		return c.SendError(msg, fmt.Errorf("cannot read cameraIndex: %s", err))
	}
	resolution, err := basic.ReadInt32(buf)
	if err != nil {
		return c.SendError(msg, fmt.Errorf("cannot read resolution: %s", err))
	}
	colorSpace, err := basic.ReadInt32(buf)
	if err != nil {
		return c.SendError(msg, fmt.Errorf("cannot read colorSpace: %s", err))
	}
	fps, err := basic.ReadInt32(buf)
	if err != nil {
		return c.SendError(msg, fmt.Errorf("cannot read fps: %s", err))
	}
	ret, callErr := p.impl.SubscribeCamera(name, cameraIndex, resolution, colorSpace, fps)

	// do not respond to post messages.
	if msg.Header.Type == net.Post {
		return nil
	}
	if callErr != nil {
		return c.SendError(msg, callErr)
	}
	var out bytes.Buffer
	errOut := basic.WriteString(ret, &out)
	if errOut != nil {
		return c.SendError(msg, fmt.Errorf("cannot write response: %s", errOut))
	}
	return c.SendReply(msg, out.Bytes())
}
func (p *stubALVideoDevice) GetImageRemote(msg *net.Message, c bus.Channel) error {
	buf := bytes.NewBuffer(msg.Payload)
	name, err := basic.ReadString(buf)
	if err != nil {
		return c.SendError(msg, fmt.Errorf("cannot read name: %s", err))
	}
	ret, callErr := p.impl.GetImageRemote(name)

	// do not respond to post messages.
	if msg.Header.Type == net.Post {
		return nil
	}
	if callErr != nil {
		return c.SendError(msg, callErr)
	}
	var out bytes.Buffer
	errOut := ret.Write(&out)
	if errOut != nil {
		return c.SendError(msg, fmt.Errorf("cannot write response: %s", errOut))
	}
	return c.SendReply(msg, out.Bytes())
}
func (p *stubALVideoDevice) Unsubscribe(msg *net.Message, c bus.Channel) error {
	buf := bytes.NewBuffer(msg.Payload)
	nameId, err := basic.ReadString(buf)
	if err != nil {
		return c.SendError(msg, fmt.Errorf("cannot read nameId: %s", err))
	}
	ret, callErr := p.impl.Unsubscribe(nameId)

	// do not respond to post messages.
	if msg.Header.Type == net.Post {
		return nil
	}
	if callErr != nil {
		return c.SendError(msg, callErr)
	}
	var out bytes.Buffer
	errOut := basic.WriteBool(ret, &out)
	if errOut != nil {
		return c.SendError(msg, fmt.Errorf("cannot write response: %s", errOut))
	}
	return c.SendReply(msg, out.Bytes())
}
func (p *stubALVideoDevice) metaObject() object.MetaObject {
	return object.MetaObject{
		Description: "ALVideoDevice",
		Methods: map[uint32]object.MetaMethod{
			100: {
				Name:                "subscribeCamera",
				ParametersSignature: "(siiii)",
				ReturnSignature:     "s",
				Uid:                 100,
			},
			101: {
				Name:                "getImageRemote",
				ParametersSignature: "(s)",
				ReturnSignature:     "m",
				Uid:                 101,
			},
			116: {
				Name:                "unsubscribe",
				ParametersSignature: "(s)",
				ReturnSignature:     "b",
				Uid:                 116,
			},
		},
		Properties: map[uint32]object.MetaProperty{},
		Signals:    map[uint32]object.MetaSignal{},
	}
}

// ALVideoDeviceProxy represents a proxy object to the service
type ALVideoDeviceProxy interface {
	SubscribeCamera(name string, cameraIndex int32, resolution int32, colorSpace int32, fps int32) (string, error)
	GetImageRemote(name string) (value.Value, error)
	Unsubscribe(nameId string) (bool, error)
	// Generic methods shared by all objectsProxy
	bus.ObjectProxy
	// WithContext can be used cancellation and timeout
	WithContext(ctx context.Context) ALVideoDeviceProxy
}

// proxyALVideoDevice implements ALVideoDeviceProxy
type proxyALVideoDevice struct {
	bus.ObjectProxy
	session bus.Session
}

// MakeALVideoDevice returns a specialized proxy.
func MakeALVideoDevice(sess bus.Session, proxy bus.Proxy) ALVideoDeviceProxy {
	return &proxyALVideoDevice{bus.MakeObject(proxy), sess}
}

// ALVideoDevice returns a proxy to a remote service
func ALVideoDevice(session bus.Session) (ALVideoDeviceProxy, error) {
	proxy, err := session.Proxy("ALVideoDevice", 1)
	if err != nil {
		return nil, fmt.Errorf("contact service: %s", err)
	}
	return MakeALVideoDevice(session, proxy), nil
}

// WithContext bound future calls to the context deadline and cancellation
func (p *proxyALVideoDevice) WithContext(ctx context.Context) ALVideoDeviceProxy {
	return MakeALVideoDevice(p.session, p.Proxy().WithContext(ctx))
}

// SubscribeCamera calls the remote procedure
func (p *proxyALVideoDevice) SubscribeCamera(name string, cameraIndex int32, resolution int32, colorSpace int32, fps int32) (string, error) {
	var err error
	var ret string
	var buf bytes.Buffer
	if err = basic.WriteString(name, &buf); err != nil {
		return ret, fmt.Errorf("serialize name: %s", err)
	}
	if err = basic.WriteInt32(cameraIndex, &buf); err != nil {
		return ret, fmt.Errorf("serialize cameraIndex: %s", err)
	}
	if err = basic.WriteInt32(resolution, &buf); err != nil {
		return ret, fmt.Errorf("serialize resolution: %s", err)
	}
	if err = basic.WriteInt32(colorSpace, &buf); err != nil {
		return ret, fmt.Errorf("serialize colorSpace: %s", err)
	}
	if err = basic.WriteInt32(fps, &buf); err != nil {
		return ret, fmt.Errorf("serialize fps: %s", err)
	}
	methodID, err := p.Proxy().MetaObject().MethodID("subscribeCamera", "(siiii)", "s")
	if err != nil {
		return ret, err
	}
	response, err := p.Proxy().CallID(methodID, buf.Bytes())
	if err != nil {
		return ret, fmt.Errorf("call subscribeCamera failed: %s", err)
	}
	resp := bytes.NewBuffer(response)
	ret, err = basic.ReadString(resp)
	if err != nil {
		return ret, fmt.Errorf("parse subscribeCamera response: %s", err)
	}
	return ret, nil
}

// GetImageRemote calls the remote procedure
func (p *proxyALVideoDevice) GetImageRemote(name string) (value.Value, error) {
	var err error
	var ret value.Value
	var buf bytes.Buffer
	if err = basic.WriteString(name, &buf); err != nil {
		return ret, fmt.Errorf("serialize name: %s", err)
	}
	methodID, err := p.Proxy().MetaObject().MethodID("getImageRemote", "(s)", "m")
	if err != nil {
		return ret, err
	}
	response, err := p.Proxy().CallID(methodID, buf.Bytes())
	if err != nil {
		return ret, fmt.Errorf("call getImageRemote failed: %s", err)
	}
	resp := bytes.NewBuffer(response)
	ret, err = value.NewValue(resp)
	if err != nil {
		return ret, fmt.Errorf("parse getImageRemote response: %s", err)
	}
	return ret, nil
}

// Unsubscribe calls the remote procedure
func (p *proxyALVideoDevice) Unsubscribe(nameId string) (bool, error) {
	var err error
	var ret bool
	var buf bytes.Buffer
	if err = basic.WriteString(nameId, &buf); err != nil {
		return ret, fmt.Errorf("serialize nameId: %s", err)
	}
	methodID, err := p.Proxy().MetaObject().MethodID("unsubscribe", "(s)", "b")
	if err != nil {
		return ret, err
	}
	response, err := p.Proxy().CallID(methodID, buf.Bytes())
	if err != nil {
		return ret, fmt.Errorf("call unsubscribe failed: %s", err)
	}
	resp := bytes.NewBuffer(response)
	ret, err = basic.ReadBool(resp)
	if err != nil {
		return ret, fmt.Errorf("parse unsubscribe response: %s", err)
	}
	return ret, nil
}

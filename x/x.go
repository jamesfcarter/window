package x

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type X struct {
	conn  *xgb.Conn
	root  xproto.Window
	atoms map[string]xproto.Atom
}

type Client struct {
	x      *X
	Window xproto.Window `json:"window"`
	Name   string        `json:"name"`
}

func New() (*X, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}
	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root
	atoms, err := internalizeAtoms(
		conn,
		root,
		"_NET_WM_NAME",
		"_NET_CLIENT_LIST",
		"_NET_ACTIVE_WINDOW",
	)
	if err != nil {
		return nil, err
	}
	return &X{
		conn:  conn,
		root:  root,
		atoms: atoms,
	}, nil
}

func (c Client) Raise() error {
	data := xproto.ClientMessageDataUnionData32New([]uint32{
		2, // source indication
		0, // timestamp
		0, // requestor's currently active window, 0 if none
		0,
		0,
	})
	ev := xproto.ClientMessageEvent{
		Format: 32,
		Window: c.Window,
		Type:   c.x.atoms["_NET_ACTIVE_WINDOW"],
		Data:   data,
	}
	return xproto.SendEventChecked(
		c.x.conn,
		true,
		c.x.root,
		xproto.EventMaskSubstructureNotify|xproto.EventMaskSubstructureRedirect,
		string(ev.Bytes()),
	).Check()
}

func internalizeAtoms(
	conn *xgb.Conn,
	root xproto.Window,
	atoms ...string,
) (map[string]xproto.Atom, error) {
	r := make(map[string]xproto.Atom)
	for _, aname := range atoms {
		reply, err := xproto.InternAtom(
			conn,
			true,
			uint16(len(aname)),
			aname,
		).Reply()
		if err != nil {
			return nil, err
		}
		r[aname] = reply.Atom
	}
	return r, nil
}

func (x *X) windowProp(w xproto.Window, prop string) (*xproto.GetPropertyReply, error) {
	return xproto.GetProperty(
		x.conn,
		false,
		w,
		x.atoms[prop],
		xproto.GetPropertyTypeAny,
		0,
		(1<<32)-1,
	).Reply()
}

func (x *X) clientList() ([]xproto.Window, error) {
	reply, err := x.windowProp(x.root, "_NET_CLIENT_LIST")
	if err != nil {
		return nil, err
	}
	count := len(reply.Value) / 4
	result := make([]xproto.Window, count)
	for i := 0; i < count; i++ {
		result[i] = xproto.Window(xgb.Get32(reply.Value[i*4:]))
	}
	return result, nil
}

func (x *X) windowName(w xproto.Window) (string, error) {
	reply, err := x.windowProp(w, "_NET_WM_NAME")
	if err != nil {
		return "", err
	}
	return string(reply.Value), nil
}

func (x *X) Clients() ([]Client, error) {
	list, err := x.clientList()
	if err != nil {
		return nil, err
	}
	r := make([]Client, len(list))
	for i, win := range list {
		r[i].x = x
		r[i].Window = win
		name, err := x.windowName(win)
		if err != nil {
			return nil, err
		}
		r[i].Name = name
	}
	return r, nil
}

package heartbeat

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

const (
	V1 uint8 = 100
)

type Handler struct {
	conn net.Conn
}

type Heartbeat struct {
	version    uint8
	dataLength uint32
	data       []byte
}

func NewHb() *Heartbeat {
	return &Heartbeat{version: V1}
}

func NewHandler(conn net.Conn) *Handler {
	resp := &Handler{conn: conn}
	return resp
}

func (h *Handler) Auth(req *AuthData) error {
	reqBytes := req.Bytes()
	//NOTE: for NODE, if you recv a AUTH msg and check it passed, you MUST REPLY ANY MSG!(e.g.OK)
	resp, err := h.Request(string(reqBytes), true)
	if err != nil {
		return err
	}
	log.Debugf("Handler.Auth resp:%v", resp)
	return nil
}

func (h *Handler) Request(data string, waitForResp bool) (resp string, err error) {
	err = NewHb().SetData(data).Write(h.conn)
	if err != nil {
		return
	}

	if !waitForResp {
		return
	}

	hbResp, err := Read(h.conn)
	if err != nil {
		return
	}
	resp = hbResp.Data()
	return
}

func SendMsg(conn net.Conn, data string) error {
	return NewHb().SetData(data).Write(conn)
}

func Read(r io.Reader) (hb *Heartbeat, err error) {
	hb = new(Heartbeat)
	if err = binary.Read(r, binary.BigEndian, &hb.version); err != nil {
		return hb, err
	}

	switch hb.version {
	case V1:
		if err = binary.Read(r, binary.BigEndian, &hb.dataLength); err != nil {
			return hb, err
		}
		if hb.dataLength != 0 {
			hb.data = make([]byte, hb.dataLength)
			if err = binary.Read(r, binary.BigEndian, &hb.data); err != nil {
				return hb, err
			}
		}
	default:
		err := fmt.Errorf("invalid version:%v", hb.version)
		log.Errorf("heartbeat.Read err:%v data:%v", err, string(hb.data))
		return nil, err
	}
	return hb, nil
}

func (hb *Heartbeat) Write(w io.Writer) error {
	return Write(w, hb)
}

// Write heartbeat
func Write(w io.Writer, hb *Heartbeat) (err error) {
	if hb == nil {
		return fmt.Errorf("%v", "hb == nil")
	}

	if _, err = w.Write([]byte{hb.version}); err != nil {
		return err
	}

	switch hb.version {
	case V1:
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, hb.dataLength)
		if _, err = w.Write(b); err != nil {
			return err
		}
		if hb.dataLength != 0 {
			if _, err = w.Write(hb.data); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("invalid version:%v", hb.version)
	}
	return nil
}

// Data get data in heartbeat
func (hb *Heartbeat) Data() string {
	return string(hb.data)
}

// SetData set data in heartbeat
func (hb *Heartbeat) SetData(data string) *Heartbeat {
	hb.data = []byte(data)
	hb.dataLength = uint32(len(hb.data))
	return hb
}

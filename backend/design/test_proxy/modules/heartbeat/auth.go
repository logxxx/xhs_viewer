package heartbeat

import (
	"encoding/json"
	"net"
)

// AuthData first connect heartbeat data
type AuthData struct {
	UserID        string `json:"user_id"`
	DeviceID      string `json:"device_id"`
	Authorization string `json:"authorization"`
	Payload       string `json:"payload"`
}

// Bytes marshal data to bytes
func (ad *AuthData) Bytes() (b []byte) {
	if ad == nil {
		return nil
	}
	b, _ = json.Marshal(ad)
	return b
}

func (ad *AuthData) DeepCopy() *AuthData {
	if ad == nil {
		return &AuthData{}
	}
	resp := &AuthData{
		UserID:        ad.UserID,
		DeviceID:      ad.DeviceID,
		Authorization: ad.Authorization,
		Payload:       ad.Payload,
	}
	return resp
}

// ParseAuthData parse from bytes
func ParseAuthData(b []byte) (ad *AuthData, err error) {
	ad = new(AuthData)
	err = json.Unmarshal(b, ad)
	return ad, err
}

func GetAuth(conn net.Conn) (*AuthData, error) {
	hb, err := Read(conn)
	if err != nil {
		return nil, err
	}
	authData, err := ParseAuthData(hb.data)
	if err != nil {
		return nil, err
	}

	err = NewHb().SetData("OK").Write(conn)
	if err != nil {
		return nil, err
	}

	return authData, nil
}

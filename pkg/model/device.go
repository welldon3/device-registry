package model

type Device struct {
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

type PageResult struct {
	Devices   []Device `json:"devices"`
	NextToken string   `json:"nextToken,omitempty"`
}

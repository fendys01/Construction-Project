package model

import (
	"panorama/bootstrap"
)

type (
	// Contract ...
	Contract struct {
		*bootstrap.App
	}
)

const (
	ChannelCustApp = "cust_mobile_app"
	ChannelCMS     = "cms"
	ChannelTCApp   = "tc_mobile_app"
)

package tappersStatus

import "github.com/up9inc/mizu/shared"

var tappersStatus map[string]*shared.TapperStatus

func Get() map[string]*shared.TapperStatus {
	if tappersStatus == nil {
		tappersStatus = make(map[string]*shared.TapperStatus)
	}

	return tappersStatus
}

func Set(tapperStatus *shared.TapperStatus) {
	if tappersStatus == nil {
		tappersStatus = make(map[string]*shared.TapperStatus)
	}

	tappersStatus[tapperStatus.NodeName] = tapperStatus
}

func Delete() {
	tappersStatus = make(map[string]*shared.TapperStatus)
}

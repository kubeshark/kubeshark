package providers

import (
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap"
	"mizuserver/pkg/models"
	"os"
	"sync"
	"time"
)

const tlsLinkRetainmentTime = time.Minute * 15

var (
	tappersCount     int
	tapStatus        *shared.TapStatus
	tappersStatus    map[string]*shared.TapperStatus
	authStatus       *models.AuthStatus
	RecentTLSLinks   = cache.New(tlsLinkRetainmentTime, tlsLinkRetainmentTime)
	tappersCountLock = sync.Mutex{}
)

func GetAuthStatus() (*models.AuthStatus, error) {
	if authStatus == nil {
		syncEntriesConfigJson := os.Getenv(shared.SyncEntriesConfigEnvVar)
		if syncEntriesConfigJson == "" {
			authStatus = &models.AuthStatus{}
			return authStatus, nil
		}

		syncEntriesConfig := &shared.SyncEntriesConfig{}
		err := json.Unmarshal([]byte(syncEntriesConfigJson), syncEntriesConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal sync entries config, err: %v", err)
		}

		if syncEntriesConfig.Token == "" {
			authStatus = &models.AuthStatus{}
			return authStatus, nil
		}

		tokenEmail, err := shared.GetTokenEmail(syncEntriesConfig.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to get token email, err: %v", err)
		}

		authStatus = &models.AuthStatus{
			Email: tokenEmail,
			Model: syncEntriesConfig.Workspace,
		}
	}

	return authStatus, nil
}

func GetAllRecentTLSAddresses() []string {
	recentTLSLinks := make([]string, 0)

	for _, outboundLinkItem := range RecentTLSLinks.Items() {
		outboundLink, castOk := outboundLinkItem.Object.(*tap.OutboundLink)
		if castOk {
			recentTLSLinks = append(recentTLSLinks, outboundLink.DstIP)
		}
	}

	return recentTLSLinks
}

func GetTapStatus() *shared.TapStatus {
	if tapStatus == nil {
		tapStatus = &shared.TapStatus{}
	}

	return tapStatus
}

func SetTapStatus(tapStatusToSet *shared.TapStatus) {
	tapStatus = tapStatusToSet
}

func GetTappersStatus() map[string]*shared.TapperStatus {
	if tappersStatus == nil {
		tappersStatus = make(map[string]*shared.TapperStatus)
	}

	return tappersStatus
}

func SetTapperStatus(tapperStatus *shared.TapperStatus) {
	if tappersStatus == nil {
		tappersStatus = make(map[string]*shared.TapperStatus)
	}

	tappersStatus[tapperStatus.NodeName] = tapperStatus
}

func DeleteTappersStatus() {
	tappersStatus = make(map[string]*shared.TapperStatus)
}

func TapperAdded() {
	tappersCountLock.Lock()
	tappersCount++
	tappersCountLock.Unlock()
}

func TapperRemoved() {
	tappersCountLock.Lock()
	tappersCount--
	tappersCountLock.Unlock()
}

func GetTappersCount() int {
	return tappersCount
}

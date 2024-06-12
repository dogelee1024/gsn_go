package common

import (
	"strings"

	trans "github.com/openwallet1/gsn-go/common/model/struct"
	"github.com/openwallet1/gsn-go/common/network"
)

type HttpHelper struct {
	relayUrl string
}

func NewHttpHelper(relayUrl string) *HttpHelper {
	return &HttpHelper{
		relayUrl: relayUrl,
	}
}

func (h *HttpHelper) AuditTransaction(request *trans.AuditRequest) (*trans.AuditResponse, error) {
	url := appendSlashTrim(h.relayUrl) + "audit"
	return network.PostWithHeader[trans.AuditResponse](url, request, nil)
}

func (h *HttpHelper) RelayTransaction(request *trans.RelayTransactionRequest) (*trans.RelayTransactionResponse, error) {
	url := appendSlashTrim(h.relayUrl) + "relay"
	return network.PostToRelay[trans.RelayTransactionResponse](url, request, nil)
}

// AppendSlashTrim trims whitespace from the input URL and ensures it ends with a slash
func appendSlashTrim(urlInput string) string {
	urlInput = strings.TrimSpace(urlInput)
	if !strings.HasSuffix(urlInput, "/") {
		urlInput += "/"
	}
	return urlInput
}

func (h *HttpHelper) GetPingResponse(relayUrl string, paymaster string) (*trans.PingResponse, error) {
	return network.GetWithHeader[trans.PingResponse](relayUrl, &trans.PintRequest{
		Paymaster: paymaster,
	}, nil)
}

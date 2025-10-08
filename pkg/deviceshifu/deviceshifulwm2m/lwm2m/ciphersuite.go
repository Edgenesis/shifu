package lwm2m

import (
	"errors"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/pion/dtls/v3"
)

// Reference:
// https://www.iana.org/assignments/tls-parameters/tls-parameters.xhtml#tls-parameters-4
// https://github.com/pion/dtls/blob/98a05d681d3affae2d055a70d3273cbb35425b5a/cipher_suite.go#L25-L45
const (
	// AES-128-CCM
	TLS_ECDHE_ECDSA_WITH_AES_128_CCM   dtls.CipherSuiteID = 0xc0ac //nolint:revive,stylecheck
	TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8 dtls.CipherSuiteID = 0xc0ae //nolint:revive,stylecheck

	// AES-128-GCM-SHA256
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 dtls.CipherSuiteID = 0xc02b //nolint:revive,stylecheck
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256   dtls.CipherSuiteID = 0xc02f //nolint:revive,stylecheck

	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 dtls.CipherSuiteID = 0xc02c //nolint:revive,stylecheck
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384   dtls.CipherSuiteID = 0xc030 //nolint:revive,stylecheck
	// AES-256-CBC-SHA
	TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA dtls.CipherSuiteID = 0xc00a //nolint:revive,stylecheck
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA   dtls.CipherSuiteID = 0xc014 //nolint:revive,stylecheck

	TLS_PSK_WITH_AES_128_CCM        dtls.CipherSuiteID = 0xc0a4 //nolint:revive,stylecheck
	TLS_PSK_WITH_AES_128_CCM_8      dtls.CipherSuiteID = 0xc0a8 //nolint:revive,stylecheck
	TLS_PSK_WITH_AES_256_CCM_8      dtls.CipherSuiteID = 0xc0a9 //nolint:revive,stylecheck
	TLS_PSK_WITH_AES_128_GCM_SHA256 dtls.CipherSuiteID = 0x00a8 //nolint:revive,stylecheck
	TLS_PSK_WITH_AES_128_CBC_SHA256 dtls.CipherSuiteID = 0x00ae //nolint:revive,stylecheck

	TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256 dtls.CipherSuiteID = 0xC037 //nolint:revive,stylecheck
)

var cipherSuiteMap = map[v1alpha1.CipherSuite]dtls.CipherSuiteID{
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM:        TLS_ECDHE_ECDSA_WITH_AES_128_CCM,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8:      TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	v1alpha1.CipherSuite_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	v1alpha1.CipherSuite_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_CCM:                TLS_PSK_WITH_AES_128_CCM,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_CCM_8:              TLS_PSK_WITH_AES_128_CCM_8,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_256_CCM_8:              TLS_PSK_WITH_AES_256_CCM_8,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_GCM_SHA256:         TLS_PSK_WITH_AES_128_GCM_SHA256,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_CBC_SHA256:         TLS_PSK_WITH_AES_128_CBC_SHA256,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	v1alpha1.CipherSuite_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	v1alpha1.CipherSuite_TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256:   TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256,
}

func CipherSuiteStringToCode(cipherSuitesStr v1alpha1.CipherSuite) (dtls.CipherSuiteID, error) {
	cipherSuiteCode, ok := cipherSuiteMap[cipherSuitesStr]
	if !ok {
		logger.Errorf("unknown cipher suite: %v", cipherSuitesStr)
		return 0, errors.New("unknown cipher suite")
	}
	return cipherSuiteCode, nil
}

func CipherSuiteStringsToCodes(cipherSuiteStrs []v1alpha1.CipherSuite) ([]dtls.CipherSuiteID, error) {
	var cipherSuiteCodes = make([]dtls.CipherSuiteID, 0, len(cipherSuiteStrs))
	for _, cipherSuiteStr := range cipherSuiteStrs {
		cipherSuiteCode, err := CipherSuiteStringToCode(cipherSuiteStr)
		if err != nil {
			return nil, err
		}
		cipherSuiteCodes = append(cipherSuiteCodes, cipherSuiteCode)
	}
	return cipherSuiteCodes, nil
}

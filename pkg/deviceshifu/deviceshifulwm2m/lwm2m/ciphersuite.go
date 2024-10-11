package lwm2m

import (
	"errors"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/pion/dtls/v2"
)

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

func StringToCode(ciperSuitStr v1alpha1.CiperSuite) (dtls.CipherSuiteID, error) {
	switch ciperSuitStr {
	case v1alpha1.CiperSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM:
		return TLS_ECDHE_ECDSA_WITH_AES_128_CCM, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8:
		return TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:
		return TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:
		return TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA, nil
	case v1alpha1.CiperSuite_TLS_PSK_WITH_AES_128_CCM:
		return TLS_PSK_WITH_AES_128_CCM, nil
	case v1alpha1.CiperSuite_TLS_PSK_WITH_AES_128_CCM_8:
		return TLS_PSK_WITH_AES_128_CCM_8, nil
	case v1alpha1.CiperSuite_TLS_PSK_WITH_AES_256_CCM_8:
		return TLS_PSK_WITH_AES_256_CCM_8, nil
	case v1alpha1.CiperSuite_TLS_PSK_WITH_AES_128_GCM_SHA256:
		return TLS_PSK_WITH_AES_128_GCM_SHA256, nil
	case v1alpha1.CiperSuite_TLS_PSK_WITH_AES_128_CBC_SHA256:
		return TLS_PSK_WITH_AES_128_CBC_SHA256, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, nil
	case v1alpha1.CiperSuite_TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256:
		return TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256, nil
	default:
		return 0, errors.New("unknown ciper suite")
	}
}

func StringsToCodes(ciperSuitStrs []v1alpha1.CiperSuite) ([]dtls.CipherSuiteID, error) {
	var ciperSuitCodes []dtls.CipherSuiteID
	for _, ciperSuitStr := range ciperSuitStrs {
		ciperSuitCode, err := StringToCode(ciperSuitStr)
		if err != nil {
			return nil, err
		}
		ciperSuitCodes = append(ciperSuitCodes, ciperSuitCode)
	}
	return ciperSuitCodes, nil
}

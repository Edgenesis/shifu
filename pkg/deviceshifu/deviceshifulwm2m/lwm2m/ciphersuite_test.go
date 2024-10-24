package lwm2m

import (
	"testing"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	"github.com/pion/dtls/v2"
)

var cipherSuiteStrs = []v1alpha1.CipherSuite{
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	v1alpha1.CipherSuite_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	v1alpha1.CipherSuite_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_CCM,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_CCM_8,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_256_CCM_8,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_GCM_SHA256,
	v1alpha1.CipherSuite_TLS_PSK_WITH_AES_128_CBC_SHA256,
	v1alpha1.CipherSuite_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	v1alpha1.CipherSuite_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	v1alpha1.CipherSuite_TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256,
}

func TestCipherSuiteStringToCode(t *testing.T) {
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

	for _, cipherSuite := range cipherSuiteStrs {
		res, err := CipherSuiteStringToCode(cipherSuite)
		if err != nil {
			t.Errorf("unknown cipher suite: %v", err)
		}
		if res != cipherSuiteMap[cipherSuite] {
			t.Errorf("Error in mapping cipher suite: %v", cipherSuite)
		}
	}
}

var want = map[dtls.CipherSuiteID]bool{
	TLS_ECDHE_ECDSA_WITH_AES_128_CCM:        true,
	TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8:      true,
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: true,
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   true,
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: true,
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   true,
	TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    true,
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      true,
	TLS_PSK_WITH_AES_128_CCM:                true,
	TLS_PSK_WITH_AES_128_CCM_8:              true,
	TLS_PSK_WITH_AES_256_CCM_8:              true,
	TLS_PSK_WITH_AES_128_GCM_SHA256:         true,
	TLS_PSK_WITH_AES_128_CBC_SHA256:         true,
	TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA256:   true,
}

func TestCipherSuiteStringsToCodes(t *testing.T) {
	res, err := CipherSuiteStringsToCodes(cipherSuiteStrs)
	if err != nil {
		t.Errorf("unknown cipher suite: %v", err)
	}

	if len(res) != len(want) {
		t.Errorf("Error in mapping cipher suite: %v", res)
	}

	for _, v := range res {
		t.Run("compare", func(t *testing.T) {
			if !compareCipherSuites(v) {
				t.Errorf("Error in mapping cipher suite")
			}
		})
	}
}

func compareCipherSuites(a dtls.CipherSuiteID) bool {
	return want[a]
}

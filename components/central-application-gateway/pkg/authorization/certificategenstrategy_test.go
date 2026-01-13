package authorization

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/application-connector-manager/components/central-application-gateway/pkg/authorization/clientcert"

	"github.com/kyma-project/kyma/application-connector-manager/components/central-application-gateway/pkg/authorization/testconsts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	certificate = []byte(testconsts.Certificate)
	privateKey  = []byte(testconsts.PrivateKey)
)

func TestCertificateGenStrategy(t *testing.T) {
	t.Run("should add certificates to proxy", func(t *testing.T) {
		// given
		clientCert := clientcert.NewClientCertificate(nil)
		certGenStrategy := newCertificateGenStrategy(certificate, privateKey)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, func(cert *tls.Certificate) {
			clientCert.SetCertificate(cert)
		}, false)
		require.NoError(t, err)

		// then
		expectedCert, err := tls.X509KeyPair(certificate, privateKey)
		require.NoError(t, err)

		assert.Equal(t, expectedCert, *clientCert.GetCertificate())
	})

	t.Run("should return error when key is invalid", func(t *testing.T) {
		// given
		certGenStrategy := newCertificateGenStrategy(certificate, []byte("invalid key"))

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, nil, false)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when certificate is invalid", func(t *testing.T) {
		// given

		certGenStrategy := newCertificateGenStrategy([]byte("invalid cert"), privateKey)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, nil, false)

		// then
		require.Error(t, err)
	})
}

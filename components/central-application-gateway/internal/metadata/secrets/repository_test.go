package secrets

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/secrets/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestRepository_Get(t *testing.T) {
	t.Run("should get given secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "CLIENT_ID", "CLIENT_SECRET", "secretId", "default-ec")
		secretsManagerMock.On("Get", context.Background(), "new-secret", metav1.GetOptions{}).Return(secret, nil)

		// when
		secrets, err := repository.Get("new-secret")

		// then
		assert.NoError(t, err)
		assert.NotNil(t, secrets["clientId"])
		assert.NotNil(t, secrets["clientSecret"])

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return an error in case fetching fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secretsManagerMock.On("Get", context.Background(), "secret-name", metav1.GetOptions{}).Return(
			nil,
			errors.New("some error"))

		// when
		cacheData, err := repository.Get("secret-name")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())
		assert.Nil(t, cacheData)

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return not found if secret does not exist", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secretsManagerMock.On("Get", context.Background(), "secret-name", metav1.GetOptions{}).Return(
			nil,
			k8serrors.NewNotFound(schema.GroupResource{},
				""))

		// when
		secrets, err := repository.Get("secret-name")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.NotEmpty(t, err.Error())

		assert.Nil(t, secrets)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("Verify caching of secret entities", func(t *testing.T) {
		// given
		secret := makeSecret("new-secret", "CLIENT_ID", "CLIENT_SECRET", "secretId", "default-ec")
		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", context.Background(), "new-secret", metav1.GetOptions{}).
			Once(). //Once() is mandatory to make this test meaningful!
			Return(secret, nil)

		repository := NewRepository(secretsManagerMock)
		require.NotNil(t, repository)

		// when
		for i := 0; i < 100; i++ {
			// only first secret will be retrieved from manager and afterwards from cache otherwise manager mock will fail
			secret, err := repository.Get("new-secret")
			assert.NoError(t, err)
			assert.Equal(t, []byte("CLIENT_ID"), secret["clientId"])
		}
	})
}

func makeSecret(name, clientID, clientSecret, serviceID, application string) *v1.Secret {
	secretMap := make(map[string][]byte)
	secretMap["clientId"] = []byte(clientID)
	secretMap["clientSecret"] = []byte(clientSecret)

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceID,
			},
		},
		Data: secretMap,
	}
}

// Package secrets contains components for accessing/modifying client secrets
package secrets

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/patrickmn/go-cache"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Repository contains operations for managing client credentials
//
//go:generate mockery --name=Repository
type Repository interface {
	Get(name string) (map[string][]byte, apperrors.AppError)
}

type repository struct {
	secretsManager Manager
	cache          *cache.Cache
	cacheRetention time.Duration
}

// Manager contains operations for managing k8s secrets
//
//go:generate mockery --name=Manager
type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Secret, error)
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManager Manager) Repository {
	cacheRetention, err := time.ParseDuration(os.Getenv("ACM_GATEWAY_SECRETCACHE_RETENTION"))
	if err != nil || cacheRetention <= 0 {
		cacheRetention = 5 * time.Minute
	}
	zap.L().Info("Configuring application cache to store application data for %.2fm", zap.Float64("cacheRetention", cacheRetention.Minutes()))
	return &repository{
		secretsManager: secretsManager,
		cache:          cache.New(cacheRetention, 3*time.Minute),
		cacheRetention: cacheRetention,
	}
}

func (r *repository) Get(name string) (map[string][]byte, apperrors.AppError) {
	cacheKey := fmt.Sprintf("secret-%s", name)
	if cachedItem, found := r.cache.Get(cacheKey); found {
		secret := cachedItem.(map[string][]byte)
		if len(secret) == 0 {
			zap.L().Warn("found empty secret '%s' in cache - this is not expected, deleting it from cache now",
				zap.String("secretName", name))
			r.cache.Delete(cacheKey)
		} else {
			return secret, nil
		}
	}

	secret, err := r.secretsManager.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		zap.L().Error("failed to read secret",
			zap.String("secretName", name),
			zap.Error(err))
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFoundf("secret '%s' not found", name)
		}
		return nil, apperrors.Internalf("failed to get '%s' secret, %s", name, err)
	}

	if err := r.cache.Add(cacheKey, secret.Data, r.cacheRetention); err != nil {
		zap.L().Warn("Failed to update secret cache entity '%s': %v", zap.String("secretName", cacheKey), zap.Error(err))
	}

	return secret.Data, nil
}

package controller

import (
	"context"

	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/logging/logger"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	gocache "github.com/patrickmn/go-cache"
	ctrlRuntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Controller interface {
	Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error)
	SetupWithManager(mgr ctrlRuntime.Manager) error
}

type controller struct {
	cacheSync CacheSync
}

func NewController(
	log *logger.Logger,
	client client.Client,
	appCache *gocache.Cache,
	appNamePlaceholder,
	eventingPathPrefixV1,
	eventingPathPrefixV2,
	eventingPathPrefixEvents string) Controller {
	return &controller{
		cacheSync: NewCacheSync(log, client, appCache, "cache_sync_controller", appNamePlaceholder, eventingPathPrefixV1, eventingPathPrefixV2, eventingPathPrefixEvents),
	}
}

func (c *controller) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return ctrlRuntime.Result{}, c.cacheSync.Sync(ctx, request.Name)
}

func (c *controller) SetupWithManager(mgr ctrlRuntime.Manager) error {

	c.cacheSync.Init(context.Background())

	return ctrlRuntime.NewControllerManagedBy(mgr).
		For(&v1alpha1.Application{}).
		Named("cache_sync_controller").
		Complete(c)
}

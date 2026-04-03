package rbac

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/kubeconfig"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

type UserCanOptions struct {
	Verb          string
	GroupResource schema.GroupResource
	Namespace     string
}

type cacheEntry struct {
	allowed bool
	expiry  time.Time
}

var (
	rbacCache = make(map[string]cacheEntry)
	mutex     = &sync.Mutex{}
)

const (
	cacheTTL = 10 * time.Second
)

func UserCan(ctx context.Context, opts UserCanOptions) (ok bool) {
	log := xcontext.Logger(ctx)

	key, err := cacheKey(ctx, opts)
	if err != nil {
		log.Error("unable to generate cache key", slog.Any("err", err))
		return false
	}

	mutex.Lock()
	entry, found := rbacCache[key]
	mutex.Unlock()

	if found && time.Now().Before(entry.expiry) {
		log.Debug("SelfSubjectAccessReviews result from cache", slog.Any("key", key))
		return entry.allowed
	}

	ep, err := xcontext.UserConfig(ctx)
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		return false
	}

	rc, err := kubeconfig.NewClientConfig(ctx, ep)
	if err != nil {
		log.Error("unable to create user client config", slog.Any("err", err))
		return false
	}

	clientset, err := kubernetes.NewForConfig(rc)
	if err != nil {
		log.Error("unable to create kubernetes clientset", slog.Any("err", err))
		return false
	}

	selfCheck := authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Group:     opts.GroupResource.Group,
				Resource:  opts.GroupResource.Resource,
				Namespace: opts.Namespace,
				Verb:      opts.Verb,
			},
		},
	}

	resp, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().
		Create(context.TODO(), &selfCheck, metav1.CreateOptions{})
	if err != nil {
		log.Error("unable to perform SelfSubjectAccessReviews",
			slog.Any("selfCheck", selfCheck), slog.Any("err", err))
		return false
	}

	log.Debug("SelfSubjectAccessReviews result", slog.Any("response", resp))

	mutex.Lock()
	rbacCache[key] = cacheEntry{
		allowed: resp.Status.Allowed,
		expiry:  time.Now().Add(cacheTTL),
	}
	mutex.Unlock()

	return resp.Status.Allowed
}

func cacheKey(ctx context.Context, opts UserCanOptions) (string, error) {
	usr, err := xcontext.User(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s:%s:%s:%s",
		usr, opts.Verb, opts.GroupResource.Group,
		opts.GroupResource.Resource, opts.Namespace), nil
}

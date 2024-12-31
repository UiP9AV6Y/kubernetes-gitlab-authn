package cache

import (
	"time"

	authentication "k8s.io/api/authentication/v1"

	ttlcache "github.com/jellydator/ttlcache/v3"
)

type UserInfoCache = ttlcache.Cache[string, authentication.UserInfo]

func NewUserInfoCache(ttl time.Duration) *UserInfoCache {
	result := ttlcache.New[string, authentication.UserInfo](
		ttlcache.WithDisableTouchOnHit[string, authentication.UserInfo](),
		ttlcache.WithTTL[string, authentication.UserInfo](ttl),
	)

	return result
}

func SetUserInfo(c *UserInfoCache, t string, u authentication.UserInfo) {
	c.Set(t, u, ttlcache.DefaultTTL)
}

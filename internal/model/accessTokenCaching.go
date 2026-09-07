package model

import (
	"fmt"
	"time"
)

// AccessTokenCacheConf holds the configuration for access token caching of a provider. At most one of the reuse
// strategies must be configured:
//   - ReuseForSeconds: a cached AT is reused while its age is at most this value
//   - ReusePercentage: a cached AT is reused while its age is at most this percentage of its lifetime
//   - ReuseIfRemainingSeconds: a cached AT is reused while it still has at least this much lifetime remaining
type AccessTokenCacheConf struct {
	ReuseForSeconds         int `yaml:"reuse_for_seconds"`
	ReusePercentage         int `yaml:"reuse_percentage"`
	ReuseIfRemainingSeconds int `yaml:"reuse_if_remaining_seconds"`
}

// Enabled returns true if access token caching is configured, i.e. at least one of the strategies is set
func (c *AccessTokenCacheConf) Enabled() bool {
	return c != nil && (c.ReuseForSeconds != 0 || c.ReusePercentage != 0 || c.ReuseIfRemainingSeconds != 0)
}

// Validate validates the access token caching configuration
func (c *AccessTokenCacheConf) Validate() error {
	if !c.Enabled() {
		return nil
	}
	numStrategies := 0
	for _, v := range []int{
		c.ReuseForSeconds,
		c.ReusePercentage,
		c.ReuseIfRemainingSeconds,
	} {
		if v < 0 {
			return fmt.Errorf("invalid values in access_token_caching: values must not be negative")
		}
		if v > 0 {
			numStrategies++
		}
	}
	if numStrategies != 1 {
		return fmt.Errorf(
			"invalid access_token_caching: exactly one of 'reuse_for_seconds', 'reuse_percentage', " +
				"'reuse_if_remaining_seconds' must be set",
		)
	}
	if c.ReusePercentage > 100 {
		return fmt.Errorf("invalid access_token_caching: 'reuse_percentage' must be at most 100")
	}
	return nil
}

// ShouldReuse returns true if an access token issued at the passed time and expiring at the passed time should be
// reused at the passed time
func (c *AccessTokenCacheConf) ShouldReuse(now, created, expiresAt time.Time) bool {
	if !c.Enabled() {
		return false
	}
	if expiresAt.IsZero() || !expiresAt.After(now) {
		return false
	}
	age := now.Sub(created)
	switch {
	case c.ReuseForSeconds > 0:
		return age.Seconds() <= float64(c.ReuseForSeconds)
	case c.ReusePercentage > 0:
		lifetime := expiresAt.Sub(created)
		if lifetime <= 0 {
			return false
		}
		return age.Seconds() <= float64(c.ReusePercentage)/100.0*lifetime.Seconds()
	case c.ReuseIfRemainingSeconds > 0:
		return expiresAt.Sub(now).Seconds() >= float64(c.ReuseIfRemainingSeconds)
	}
	return false
}

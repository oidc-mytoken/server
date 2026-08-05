package model

import (
	"testing"
	"time"
)

func TestAccessTokenCacheConf_Enabled(t *testing.T) {
	tests := []struct {
		name string
		conf *AccessTokenCacheConf
		want bool
	}{
		{
			"nil",
			nil,
			false,
		},
		{
			"empty",
			&AccessTokenCacheConf{},
			false,
		},
		{
			"reuse_for_seconds",
			&AccessTokenCacheConf{ReuseForSeconds: 300},
			true,
		},
		{
			"reuse_percentage",
			&AccessTokenCacheConf{ReusePercentage: 30},
			true,
		},
		{
			"reuse_if_remaining_seconds",
			&AccessTokenCacheConf{ReuseIfRemainingSeconds: 60},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.conf.Enabled(); got != tt.want {
					t.Errorf("Enabled() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestAccessTokenCacheConf_Validate(t *testing.T) {
	tests := []struct {
		name    string
		conf    *AccessTokenCacheConf
		wantErr bool
	}{
		{
			"nil",
			nil,
			false,
		},
		{
			"empty",
			&AccessTokenCacheConf{},
			false,
		},
		{
			"single strategy",
			&AccessTokenCacheConf{ReuseForSeconds: 300},
			false,
		},
		{
			"two strategies",
			&AccessTokenCacheConf{
				ReuseForSeconds: 300,
				ReusePercentage: 30,
			},
			true,
		},
		{
			"all strategies",
			&AccessTokenCacheConf{
				ReuseForSeconds:         300,
				ReusePercentage:         30,
				ReuseIfRemainingSeconds: 60,
			},
			true,
		},
		{
			"negative",
			&AccessTokenCacheConf{ReuseForSeconds: -5},
			true,
		},
		{
			"percentage over 100",
			&AccessTokenCacheConf{ReusePercentage: 101},
			true,
		},
		{
			"percentage 100",
			&AccessTokenCacheConf{ReusePercentage: 100},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.conf.Validate()
				if (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestAccessTokenCacheConf_ShouldReuse(t *testing.T) {
	now := time.Now()
	created := now.Add(-10 * time.Second)  // AT is 10s old
	expiresAt := now.Add(90 * time.Second) // 90s remaining, 100s lifetime

	tests := []struct {
		name string
		conf *AccessTokenCacheConf
		want bool
	}{
		// reuse_for_seconds: reuse while age <= 5s -> 10s old is too old
		{
			"for_seconds too old",
			&AccessTokenCacheConf{ReuseForSeconds: 5},
			false,
		},
		// reuse_for_seconds: reuse while age <= 15s -> 10s old is ok
		{
			"for_seconds ok",
			&AccessTokenCacheConf{ReuseForSeconds: 15},
			true,
		},
		// reuse_percentage: reused for first 20% of 100s lifetime (=20s) -> 10s ok
		{
			"percentage ok",
			&AccessTokenCacheConf{ReusePercentage: 20},
			true,
		},
		// reuse_percentage: reused for first 5% of 100s lifetime (=5s) -> 10s too old
		{
			"percentage too old",
			&AccessTokenCacheConf{ReusePercentage: 5},
			false,
		},
		// reuse_if_remaining_seconds: reuse while remaining >= 120s -> only 90s left
		{
			"remaining too low",
			&AccessTokenCacheConf{ReuseIfRemainingSeconds: 120},
			false,
		},
		// reuse_if_remaining_seconds: reuse while remaining >= 60s -> 90s left ok
		{
			"remaining ok",
			&AccessTokenCacheConf{ReuseIfRemainingSeconds: 60},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.conf.ShouldReuse(now, created, expiresAt); got != tt.want {
					t.Errorf("ShouldReuse() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestAccessTokenCacheConf_ShouldReuse_Expired(t *testing.T) {
	now := time.Now()
	created := now.Add(-10 * time.Second)
	expiresAt := now.Add(-1 * time.Second) // already expired
	conf := &AccessTokenCacheConf{ReuseForSeconds: 60}
	if conf.ShouldReuse(now, created, expiresAt) {
		t.Error("ShouldReuse() must return false for an expired token")
	}
}

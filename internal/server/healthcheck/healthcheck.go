package healthcheck

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/httpclient"
	"github.com/oidc-mytoken/utils/utils"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/go-oidfed/pkg"
	"github.com/zachmann/go-oidfed/pkg/cache"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/versionrepo"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

// Start starts the healthcheck endpoint on the configured port if enabled
func Start() {
	if !config.Get().Server.Healthcheck.Enabled {
		return
	}
	httpServer := fiber.New()
	httpServer.Get(
		"", handleHealthCheck,
	)
	addr := fmt.Sprintf(":%d", config.Get().Server.Healthcheck.Port)
	log.Infof("Healthcheck endpoint started on %s", addr)
	go func() {
		log.WithError(httpServer.Listen(addr)).Fatal()
	}()
}

type status struct {
	Healthy     bool             `json:"healthy"`
	Operational bool             `json:"operational"`
	Components  componentsStatus `json:"components"`
	Version     string           `json:"version"`
	Timestamp   pkg.Unixtime     `json:"timestamp"`
}

type componentsStatus struct {
	ServerUp        bool `json:"server_up"`
	ServerReachable bool `json:"server_reachable"`
	Database        bool `json:"database_up"`
	Cache           bool `json:"cache_up"`
}

func (c componentsStatus) healthy() bool {
	return c.ServerUp && c.ServerReachable && c.Database && c.Cache
}
func (c componentsStatus) operational() bool {
	return c.ServerUp && c.ServerReachable && c.Database
}

func handleHealthCheck(ctx *fiber.Ctx) error {
	state := healthcheck()
	if !state.Operational {
		ctx.Status(fiber.StatusServiceUnavailable)
	}
	return ctx.JSON(state)
}

func healthcheck() status {
	components := componentsStatus{
		ServerUp:        true,
		ServerReachable: checkServer(),
		Database:        checkDB(),
		Cache:           checkCache(),
	}
	return status{
		Healthy:     components.healthy(),
		Operational: components.operational(),
		Components:  components,
		Version:     version.VERSION,
		Timestamp:   pkg.Unixtime{Time: time.Now()},
	}
}

func checkServer() bool {
	_, err := httpclient.Do().R().Get(routes.ConfigEndpoint)
	if err != nil {
		log.WithError(err).WithField("healthcheck", "server_reachable").Error("error server healthcheck")
		return false
	}
	return true
}

func checkDB() bool {
	_, err := versionrepo.GetVersionState(log.StandardLogger(), nil)
	if err != nil {
		log.WithError(err).WithField("healthcheck", "db").Error("error db healthcheck")
		return false
	}
	return true
}

var cacheMutex sync.Mutex

func checkCache() bool {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	k := "healthcheck"
	v := utils.RandASCIIString(64)
	if err := cache.Set(k, v, time.Second); err != nil {
		log.WithError(err).WithField("healthcheck", "cache").Error("error caching healthcheck")
		return false
	}
	var cached string
	set, err := cache.Get(k, &cached)
	if err != nil {
		log.WithError(err).WithField("healthcheck", "cache").
			Error("error obtaining cached healthcheck")
		return false
	}
	if !set {
		log.WithField("healthcheck", "cache").Error("cached healthcheck not found")
		return false
	}
	if cached != v {
		log.WithField("healthcheck", "cache").
			WithField("cached", v).
			WithField("obtained", cached).
			Error("cached value does not match")
		return false
	}
	return true
}

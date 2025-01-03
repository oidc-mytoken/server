package server

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/profilerepo"
	configurationEndpoint "github.com/oidc-mytoken/server/internal/endpoints/configuration"
	"github.com/oidc-mytoken/server/internal/endpoints/webentities"
	"github.com/oidc-mytoken/server/internal/utils/cache"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/templating"
)

func handleIndex(ctx *fiber.Ctx) error {
	binding := homeBindingData()
	binding[templating.MustacheKeyLoggedIn] = false
	binding[templating.MustacheKeyCookieLifetime] = cookies.CookieLifetime
	return ctx.Render("sites/home", binding, templating.LayoutMain)
}

func homeBindingData() map[string]interface{} {
	var providers []map[string]any
	for _, p := range configurationEndpoint.SupportedProviders() {
		pp := make(map[string]any, 2)
		pp["issuer"] = p.Issuer
		pp["name"] = p.Name
		pp["fed"] = p.OIDCFed
		providers = append(providers, pp)
	}
	bindingData := map[string]interface{}{
		templating.MustacheKeyLoggedIn:        true,
		templating.MustacheKeyRestrictionsGUI: true,
		templating.MustacheKeyHome:            true,
		templating.MustacheKeyCapabilities:    webentities.AllWebCapabilities(),
		templating.MustacheSubTokeninfo: map[string]interface{}{
			templating.MustacheKeyCollapse: templating.Collapsable{
				CollapseRestr: true,
			},
			templating.MustacheKeyPrefix:            "tokeninfo-",
			templating.MustacheKeyReadOnly:          true,
			templating.MustacheKeyCalendarsEditable: false,
		},
		templating.MustacheSubMTListing: map[string]interface{}{
			templating.MustacheKeySubscribeNotifications: true,
		},
		templating.MustacheSubCreateMT: map[string]interface{}{
			templating.MustacheKeyPrefix:             "createMT-",
			templating.MustacheKeyCreateWithProfiles: true,
			templating.MustacheKeyProfiles:           profilesBindingData(),
		},
		"providers": providers,
	}
	if config.Get().Features.Notifications.ICS.Enabled || config.Get().
		Features.Notifications.Mail.Enabled {
		bindingData[templating.MustacheSubNotifications] = map[string]interface{}{
			templating.MustacheKeyPrefix:              "notifications-",
			templating.MustacheKeyNotificationClasses: webentities.AllWebNotificationClass(),
			"modify": map[string]any{
				templating.MustacheKeyPrefix: "notifications-modify-",
			},
			templating.MustacheSubNotificationListing: map[string]any{
				templating.MustacheKeyPrefix: "notification-listing-",
				templating.MustacheSubNewNotificationModal: map[string]any{
					templating.MustacheKeyPrefix: "new-notification-modal-",
				},
			},
		}
		bindingData[templating.MustacheKeyNotificationsMailEnabled] = config.Get().Features.Notifications.Mail.Enabled
		bindingData[templating.MustacheKeyNotificationsCalendarEnabled] = config.Get().Features.Notifications.ICS.Enabled
	}
	return bindingData
}

type templateProfileData struct {
	Name    string
	Payload string
}

// getWebProfileData returns the cached profile data for one of the profile types
func getWebProfileData(t string) (data []templateProfileData, ok bool) {
	found, err := cache.Get(cache.WebProfiles, t, &data)
	if err != nil {
		log.WithError(err).Error("failed to fetch web profile data")
	}
	if err != nil || !found {
		ok = false
	}
	return
}

func profilesBindingData() map[string]interface{} {
	var groups []string

	groupsFound, err := cache.Get(cache.WebProfiles, "groups", &groups)
	if err != nil || !groupsFound {
		groups, err = profilerepo.GetGroups(log.StandardLogger(), nil)
		if err != nil {
			log.WithError(err).Error("error while retrieving profile groups for webinterface binding data")
			return nil
		}
		if err = cache.Set(
			cache.WebProfiles, "groups", groups,
			time.Hour,
		); err != nil {
			log.WithError(err).Error("error while setting profile groups for webinterface binding data")
		}
	}

	profileTypes := map[string]func(log.Ext1FieldLogger, *sqlx.Tx, string) (profiles []api.Profile, err error){
		templating.MustacheKeyProfilesProfile:      profilerepo.GetProfiles,
		templating.MustacheKeyProfilesCapabilities: profilerepo.GetCapabilitiesTemplates,
		templating.MustacheKeyProfilesRestrictions: profilerepo.GetRestrictionsTemplates,
		templating.MustacheKeyProfilesRotation:     profilerepo.GetRotationTemplates,
	}
	returnData := make(map[string]interface{})
	for pt, dbFunc := range profileTypes {
		p, ok := getWebProfileData(pt)
		if ok {
			returnData[pt] = p
			continue
		}
		p = []templateProfileData{}
		for _, group := range groups {
			profilesForGroup, err := dbFunc(log.StandardLogger(), nil, group)
			if err != nil {
				log.WithError(err).WithField("profile type", pt).Error(
					"error while retrieving profiles for webinterface binding data",
				)
			}
			for _, d := range profilesForGroup {
				payload, err := d.Payload.MarshalJSON()
				if err != nil {
					log.WithError(err).WithField("profile type", pt).Error(
						"error while marshaling payload while retrieving profiles for webinterface" +
							" binding data",
					)
				}
				p = append(
					p, templateProfileData{
						Name:    fmt.Sprintf("%s/%s", group, d.Name),
						Payload: string(payload),
					},
				)
			}
		}
		sort.Slice(
			p, func(i, j int) bool {
				return p[i].Name < p[j].Name
			},
		)
		for i, pp := range p {
			pp.Name = strings.TrimPrefix(pp.Name, "_/")
			p[i] = pp
		}
		returnData[pt] = p
		cache.Set(cache.WebProfiles, pt, p, time.Hour)
	}
	return returnData
}

func handleHome(ctx *fiber.Ctx) error {
	return ctx.Render("sites/home", homeBindingData(), templating.LayoutMain)
}

func handleViewCalendar(ctx *fiber.Ctx) error {
	return ctx.Render(
		"sites/calendar", map[string]any{
			"calendar-view":                   true,
			templating.MustacheKeyEmptyNavbar: true,
		}, templating.LayoutMain,
	)
}

func handleSettings(ctx *fiber.Ctx) error {
	type bindData struct {
		DisplayName string
		Name        string
		Description string
		Link        string
		EmbedBody   string
		partialName string
		bindingData map[string]interface{}
	}
	grants := []*bindData{
		{
			DisplayName: "SSH",
			Name:        "ssh",
			Description: "The SSH grant type allows you to link an ssh key and use ssh authentication for various actions.",
			Link:        "/settings/grants/ssh",
			partialName: "sites/settings-ssh",
			bindingData: map[string]interface{}{
				templating.MustacheKeyRestrictionsGUI: true,
				templating.MustacheKeyRestrictions:    webentities.WebRestrictions{},
				templating.MustacheKeyCapabilities:    webentities.AllWebCapabilities(),
			},
		},
	}
	for _, g := range grants {
		embed := strings.Builder{}
		if err := errors.WithStack(serverConfig.Views.Render(&embed, g.partialName, g.bindingData)); err != nil {
			return err
		}
		g.EmbedBody = embed.String()
	}
	binding := map[string]interface{}{
		templating.MustacheKeyGrants:            grants,
		templating.MustacheKeyLoggedIn:          true,
		templating.MustacheKeySettings:          true,
		templating.MustacheKeySettingsSSH:       true,
		templating.MustacheKeyRestrictionsGUI:   true,
		templating.MustacheKeyCalendarsEditable: true,
		templating.MustacheKeyRestrictions:      webentities.WebRestrictions{},
		templating.MustacheKeyCapabilities:      webentities.AllWebCapabilities(),
		templating.MustacheKeyPrefix:            "settings-",
	}
	if config.Get().Features.Notifications.ICS.Enabled || config.Get().
		Features.Notifications.Mail.Enabled {
		binding[templating.MustacheSubNotifications] = true
		binding[templating.MustacheKeyNotificationsMailEnabled] = config.Get().Features.Notifications.Mail.Enabled
		binding[templating.MustacheKeyNotificationsCalendarEnabled] = config.Get().Features.Notifications.ICS.Enabled
	}
	return ctx.Render("sites/settings", binding, templating.LayoutMain)
}

func handleNativeCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		templating.MustacheKeyEmptyNavbar: true,
		templating.MustacheKeyApplication: ctx.Query("application"),
	}
	return ctx.Render("sites/native", binding, templating.LayoutMain)
}

func handleNativeConsentAbortCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		templating.MustacheKeyEmptyNavbar: true,
		templating.MustacheKeyApplication: ctx.Query("application"),
	}
	return ctx.Render("sites/native.abort", binding, templating.LayoutMain)
}

func handlePrivacy(ctx *fiber.Ctx) error {
	so := config.Get().ServiceOperator
	binding := map[string]interface{}{
		templating.MustacheKeyEmptyNavbar:    true,
		templating.MustacheKeyName:           so.Name,
		templating.MustacheKeyHomepage:       so.Homepage,
		templating.MustacheKeyContact:        so.Contact,
		templating.MustacheKeyPrivacyContact: so.Privacy,
	}
	return ctx.Render("sites/privacy", binding, templating.LayoutMain)
}

func handleNotificationManagement(ctx *fiber.Ctx) error {
	return ctx.Render(
		"sites/manage-notification", map[string]any{
			"notification-management":                  true,
			"empty-navbar":                             true,
			templating.MustacheSubNewNotificationModal: true,
			templating.MustacheKeyNotificationClasses:  webentities.AllWebNotificationClass(),
			templating.MustacheKeyCollapse: map[string]bool{
				"NotificationManagement": true,
			},
		}, templating.LayoutMain,
	)
}

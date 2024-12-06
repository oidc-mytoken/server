package webentities

import (
	"strings"

	"github.com/oidc-mytoken/api/v0"
)

// WebNotificationClass is type for representing api.NoticationClass in the web compatible with WebCapability
type WebNotificationClass struct {
	ReadWriteCapability *api.NotificationClass
	Children            []*WebNotificationClass
}

// WebNotificationClasses creates a slice of WebNotificationClass from []api.NotificationClass
func WebNotificationClasses(ncs []*api.NotificationClass) (wnc []*WebNotificationClass) {
	for _, nc := range ncs {
		wnc = append(
			wnc, webNotificationClassFromNotificationClass(nc),
		)
	}
	return
}

// AllWebNotificationClass returns all WebNotificationClass as a tree
func AllWebNotificationClass() []*WebNotificationClass {
	return allWebNotificationClass
}

var allWebNotificationClass []*WebNotificationClass

func init() {
	if allWebNotificationClass == nil {
		allWebNotificationClass = []*WebNotificationClass{}
	}
	for _, nc := range api.AllNotificationClasses {
		if strings.Contains(nc.Name, ":") {
			continue
		}
		allWebNotificationClass = append(
			allWebNotificationClass, notificationClassToWebNotificationClass(nc),
		)
	}
}

func notificationClassToWebNotificationClass(nc *api.NotificationClass) *WebNotificationClass {
	var childs []*WebNotificationClass
	for _, c := range nc.GetChildren() {
		childs = append(childs, notificationClassToWebNotificationClass(c))
	}
	return &WebNotificationClass{
		ReadWriteCapability: nc,
		Children:            childs,
	}
}

func webNotificationClassFromNotificationClass(nc *api.NotificationClass) *WebNotificationClass {
	for _, wnc := range allWebNotificationClass {
		if wnc.ReadWriteCapability.Name == nc.Name {
			return wnc
		}
	}
	return nil
}

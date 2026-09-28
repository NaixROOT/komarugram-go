// SPDX-License-Identifier: Unlicense OR MIT

package services

var (
	appNotificationSingleton AppNotification
)

type AppNotification interface {
	ShowNotification(msg string)
}

func GetNotificationService() AppNotification {
	return appNotificationSingleton
}

func SetNotificationService(appNotifications AppNotification) {
	appNotificationSingleton = appNotifications
}

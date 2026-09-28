// SPDX-License-Identifier: Unlicense OR MIT

package services

import "komarugram/internal/motion"

var (
	appMotionSingleton *motion.Settings
)

func GetMotionService() *motion.Settings {
	return appMotionSingleton
}

func SetMotionService(appMotion *motion.Settings) {
	appMotionSingleton = appMotion
}

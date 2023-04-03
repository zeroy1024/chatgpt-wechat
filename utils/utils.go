package utils

import "regexp"

func IsValidUUIDv4(uuid string) bool {
	uuidV4Re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuid != "" && uuidV4Re.MatchString(uuid)
}

package service

import "strconv"

func StringToUint(s string) uint {
	if array, err := strconv.ParseUint(s, 10, 64); err == nil {
		return uint(array)
	}
	return 0
}

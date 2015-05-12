package doaj

import "strconv"

func IsInteger(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

func IsYear(s string) bool {
	v, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	if v < 1 || v > 9999 {
		return false
	}
	return true
}

func IsMonth(s string) bool {
	v, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	if v < 1 || v > 12 {
		return false
	}
	return true
}

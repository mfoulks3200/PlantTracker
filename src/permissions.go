package main

import "strings"

func comparePermissions(requiredPerm string, comparedPerm string) bool {
	if comparedPerm == "*" {
		return true
	}

	requiredPermTree := strings.Split(requiredPerm, ".")
	comparedPermTree := strings.Split(comparedPerm, ".")

	for i := 0; i < len(requiredPermTree); i++ {
		if comparedPermTree[i] == "*" {
			return true
		} else if comparedPermTree[i] == requiredPermTree[i] && i == len(requiredPermTree) {
			return true
		} else if comparedPermTree[i] != requiredPermTree[i] {
			return false
		}
	}
	return false
}

func checkPermissions(user User, permission string) bool {
	if permission == "" {
		return true
	}
	for i := 0; i < len(user.Permissions); i++ {
		if comparePermissions(permission, user.Permissions[i]) {
			return true
		}
	}
	return false
}

package main

import "strings"

const pathSeparator = "/"

func replaceSlash(path string) string {
	return strings.ReplaceAll(path, pathSeparator, "-")
}

func addTrailingSlash(path string) string {
	if path[len(path)-1:] == pathSeparator {
		return path
	}
	return path + pathSeparator
}

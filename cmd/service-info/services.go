package main

import (
	"sort"
)

var servicePorts = map[string]int{
	"api-server": 8090,
}

func allServices() []string {
	result := make([]string, 0, len(servicePorts))
	for name := range servicePorts {
		result = append(result, name)
	}
	sort.Strings(result)

	return result
}
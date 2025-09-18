package main

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/registry/helpers"
)

func main() {
	testCases := []string{
		"://invalid-url",
		"https://registry..example.com",
	}
	
	for _, test := range testCases {
		result, err := helpers.NormalizeRegistryURL(test)
		fmt.Printf("Input: %q -> Result: %q, Error: %v\n", test, result, err)
	}
}

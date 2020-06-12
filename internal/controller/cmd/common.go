package main

import "fmt"

func namespacedPodName(namespace, name string) string {
	return fmt.Sprintf("%s:%s", namespace, name)
}

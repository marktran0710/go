package main

import (
	route "github.com/marktran77/go/router"
)

func main() {
	var router = route.SetUpRouter()
	router.Run("localhost:8080")
}

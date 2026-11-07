package main

import (
	"fmt"
	"url-shortener/internal/config"
)

func main() {

	cfg := config.MustLoadConfig()

	fmt.Println(cfg)
}
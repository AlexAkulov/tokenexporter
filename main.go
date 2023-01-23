package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	listen := os.Getenv("LISTEN")
	configFile := os.Getenv("CONFIG_FILE")

	config, err := ReadConfig(configFile)
	if err != nil {
		log.Println("can't read config:", err)
		os.Exit(1)
	}

	watcher := Watcher{
		geth:  map[string]*ethclient.Client{},
		items: map[string][]*watchItem{},
	}
	if err := watcher.LoadConfig(config); err != nil {
		log.Println("can't load config:", err)
		os.Exit(1)
	}
	if err := watcher.Start(listen); err != nil {
		log.Println("can't start watcher:", err)
		os.Exit(1)
	}
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)
	log.Println("Started")
	<- sigterm
	log.Println("Stop")
}

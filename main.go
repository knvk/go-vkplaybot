package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"os/signal"
	"syscall"
	"vkplaybot/vkplaybot"
)

func main() {

	if len(os.Args) < 2 {
		log.Println("usage: bot <config>")
		os.Exit(0)
	}
	f := os.Args[1]
	if _, err := os.Stat(f); err != nil {
		log.Fatal(err)
	}

	config := &vkplaybot.BotConfig{}
	_, err := toml.DecodeFile(f, config)
	if err != nil {
		log.Fatal(err)
	}
	bot := &vkplaybot.VKPlayBot{
		Config: *config,
	}

	err = bot.Start()

	if err != nil {
		log.Fatal(err)
	}

	bot.Handle()
	bot.GetViewers()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		bot.Cancel <- struct{}{}
		bot.SendChatMessage("bye", bot.GetChannel(), nil)
		//time.Sleep(1* time.Second)
		os.Exit(1)
	}()
	select {}
}

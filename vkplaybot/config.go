package vkplaybot

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	GetChannelURL string = "https://api.vkplay.live/v1/blog/"
)

type BotConfig struct {
	User       string
	Pass       string
	Token      string
	Channel    string
	Modules    []string
	BanPhrases string
}

func (cfg *BotConfig) getBanPhrases() ([]string, error) {
	file, err := os.Open(cfg.BanPhrases)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// get channels name and returns Channel struct for bot
func (cfg *BotConfig) getChannelsFromConfig() (*Channel, error) {
	r, _ := http.NewRequest("GET", GetChannelURL+cfg.Channel, nil)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ch := &Channel{}
	json.Unmarshal(b, &ch)

	return ch, nil
}

// activates and append module to slice of
// order matter
func (cfg *BotConfig) activateModules() ([]Module, error) {
	modules := []Module{}
	for i := 0; i < len(cfg.Modules); i++ {
		switch cfg.Modules[i] {
		case "commands":
			cm := NewCommandsModule()
			modules = append(modules, cm)
		case "filter":
			bw, _ := cfg.getBanPhrases()
			fm := NewFilterModule(bw)
			modules = append(modules, fm)
		case "poll":
			pm := NewPollModule()
			modules = append(modules, pm)
		default:
			log.Printf("unknown module %s skipping ", cfg.Modules[i])
		}
	}
	return modules, nil
}

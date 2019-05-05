package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/jamesfcarter/window/x"
)

func onClients(clients []x.Client, f func(x.Client)) {
	for _, client := range clients {
		f(client)
	}
}

func filterClients(clients []x.Client, filter string) ([]x.Client, error) {
	re, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	r := make([]x.Client, 0, len(clients))
	for _, client := range clients {
		if !re.MatchString(client.Name) {
			continue
		}
		r = append(r, client)
	}
	return r, nil
}

func filterClientsID(clients []x.Client, filter uint) ([]x.Client, error) {
	win := xproto.Window(filter)
	r := make([]x.Client, 0, 1)
	for _, client := range clients {
		if client.Window != win {
			continue
		}
		r = append(r, client)
	}
	if len(r) == 0 {
		return nil, fmt.Errorf("no such window 0x%08x", filter)
	}
	return r, nil
}

func main() {
	log.SetFlags(0)

	var jsonFlag bool
	flag.BoolVar(&jsonFlag, "json", false, "output window list as JSON")
	var listFlag bool
	flag.BoolVar(&listFlag, "list", false, "output window list")
	var filter string
	flag.StringVar(&filter, "filter", "", "regex to filter window list")
	var filterID uint
	flag.UintVar(&filterID, "id", 0, "select a specific window")
	var raiseFlag bool
	flag.BoolVar(&raiseFlag, "raise", false, "raise the selected windows")

	flag.Parse()

	var err error
	X, err := x.New()
	if err != nil {
		log.Fatal(err)
	}

	clients, err := X.Clients()
	if err != nil {
		log.Fatal(err)
	}

	if filter != "" {
		clients, err = filterClients(clients, filter)
		if err != nil {
			log.Fatal(err)
		}
	}
	if filterID > 0 {
		clients, err = filterClientsID(clients, filterID)
		if err != nil {
			log.Fatal(err)
		}
	}

	if raiseFlag {
		onClients(clients, func(client x.Client) {
			err := client.Raise()
			if err != nil {
				log.Fatal(err)
			}
		})
	}

	if jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(clients)
	}
	if listFlag {
		for _, client := range clients {
			fmt.Printf("0x%08x %s\n", client.Window, client.Name)
		}
	}
}

package main

import (
	"fmt"
	pavarotti "github.com/viking/go-pavarotti"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Syntax: %s <path>\n", os.Args[0])
		os.Exit(1)
	}

	library, err := pavarotti.NewLibrary("pavarotti.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ch := make(chan pavarotti.Song)
	quit := make(chan bool)
	go func() {
		for {
			select {
			case song := <-ch:
				fmt.Printf("%+v\n", song)
				library.AddSong(song)
			case <-quit:
				return
			}
		}
	}()
	err = pavarotti.Find(os.Args[1], ch)
	quit <- true

	if err == nil {
		err = library.Close()
	}

	if err != nil {
		fmt.Println(err)
	}
}

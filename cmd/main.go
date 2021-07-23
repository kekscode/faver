package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/kekscode/faver/internal"
)

func main() {
	flag.Parse()

	// read program input
	var targets []string
	if flag.NArg() == 0 { // from stdin/pipe
		fmt.Println("Reading from stdin")

		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			log.Println("line", s.Text())
			targets = append(targets, s.Text())
		}
	} else { // from argument
		fmt.Println("Reading positional arguments")
		targets = flag.Args()
	}

	var wg sync.WaitGroup
	wg.Add(len(targets))
	for _, arg := range targets {
		go func(arg string) {
			defer wg.Done()
			log.Printf("Starting goroutine for %v", arg)
			// Find and fetch favicon:
			f := internal.New()
			favicons, err := f.FetchFavicons(arg)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			if len(favicons) == 0 {
				log.Fatalf("error: no favicons for %s found", arg)
				return
			}

			for idxIco, icon := range favicons {
				u, err := url.Parse(
					arg,
				)
				if err != nil {
					log.Fatal(err)
				}

				// Save to disk:
				curtime := time.Now()
				fname := fmt.Sprintf("%s-%s-%d.ico",
					fmt.Sprintf(
						"%d%02d%02d%02d%02d%02d",
						curtime.Year(), curtime.Month(), curtime.Day(),
						curtime.Hour(), curtime.Minute(), curtime.Second(),
					),
					u.Host,
					idxIco)
				out, err := os.Create(fname)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				defer out.Close()

				_, err = out.Write(icon)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
			}
		}(arg)
	}
	wg.Wait()
}

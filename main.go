package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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

	for _, arg := range targets {
		// Find and fetch favicon:
		favicons, err := FetchFavicons(arg)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if len(favicons) == 0 {
			log.Fatalf("error: no favicons for %s found", arg)
			break
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
	}

}

// FetchFavicon fetches favicon data
func FetchFavicons(url string) (icons [][]byte, err error) {
	iconsURL, err := findFavicons(url)
	if err != nil {
		return nil, err
	}

	// download favicons
	var iconsData [][]byte
	for _, ico := range iconsURL {
		icoData, err := getFavicon(ico)
		if err != nil {
			return nil, err
		}
		iconsData = append(iconsData, icoData)
	}

	return iconsData, nil
}

// Go to loc, follow redirects, download html,
// parse body for <link rel="icon" href="path-to-icon">
// and return all paths to referenced favicons
// Also: blindly try favicon.ico in the site root
func findFavicons(loc string) ([]string, error) {
	resp, err := http.Get(loc)
	if err != nil {
		return []string{""}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return []string{""}, err
	}

	var icoHrefs []string
	var found bool

	doc.Find("head link[rel=icon]").Each(func(i int, s *goquery.Selection) {
		href := ""
		href, found = s.Attr("href")

		icoHrefs = append(icoHrefs, href)
		if !found {
			log.Println("cannot find a favicon URL in HTML body")
		}
	})

	var favicons []string
	for _, ico := range icoHrefs {
		if strings.HasPrefix(ico, "/") {
			// Relative path found, build a full qualified path
			u, err := url.Parse(
				resp.Request.URL.Scheme + "://" +
					resp.Request.URL.Host +
					ico,
			)

			if err != nil {
				return nil, err
			}
			favicons = append(favicons, u.String())
		}
	}

	// Last resort: Try favicon.ico in document root
	if len(favicons) == 0 {
		docroot := loc + "/favicon.ico"
		resp, err := http.Get(docroot)
		if err != nil {
			log.Fatalf("Cannot find %v\n", err)
			return nil, err
		}
		defer resp.Body.Close()

		favicons = append(favicons, docroot)
	}

	return favicons, nil
}

func getFavicon(url string) (icon []byte, err error) {
	if len(url) >= 7 {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		icon, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return icon, nil
	} else {
		return nil, fmt.Errorf("URL \"%v\" is not valid", url)
	}
}

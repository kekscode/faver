package internal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Fetcher struct{}

func New() *Fetcher {
	f := Fetcher{}
	return &f
}

// FetchFavicons fetches favicon data for a given url
func (f *Fetcher) FetchFavicons(url string) (icons [][]byte, err error) {
	iconsURL, err := f.findFavicons(url)
	if err != nil {
		return nil, err
	}

	// download favicons
	var iconsData [][]byte
	for _, ico := range iconsURL {
		icoData, err := f.getFavicon(ico)
		if err != nil {
			return nil, err
		}
		iconsData = append(iconsData, icoData)
	}

	return iconsData, nil
}

// findFavicons tries to find favicon URLs for a given location
func (f *Fetcher) findFavicons(loc string) ([]string, error) {
	// Go to loc, follow redirects, download html,
	// parse body for <link rel="icon" href="path-to-icon">
	// and return all paths to referenced favicons
	// Also: blindly try favicon.ico in the site root
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

// getFavicon downloads a favicon
func (f *Fetcher) getFavicon(url string) (icon []byte, err error) {
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

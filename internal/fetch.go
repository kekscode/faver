/*
Internal Fetcher object
*/
package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"crypto/tls"

	"github.com/PuerkitoBio/goquery"
)

type Fetcher struct{}

func New() *Fetcher {
	f := Fetcher{}
	return &f
}

// FetchFavicons fetches favicon data for a given url
func (f *Fetcher) FetchFavicons(url string) (data [][]byte, err error) {
	iconsURL, err := f.findFavicons(url)
	if err != nil {
		return nil, err
	}

	// download favicons
	for _, ico := range iconsURL {
		icoData, err := f.getFavicon(ico)
		if err != nil {
			return nil, err
		}
		data = append(data, icoData)
	}

	return data, nil
}

// getHTML downloads raw HTML and request data for an URL
func (f *Fetcher) getHTML(url string) (body io.Reader, response *http.Response, err error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}
	response, err = client.Get(url)
	if err != nil {
		return nil, response, err
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, response, err
	}

	html := io.NopCloser(bytes.NewBuffer(b))

	return html, response, nil
}

// findFavicons tries to find favicon URLs for a given location
func (f *Fetcher) findFavicons(loc string) (icons []string, err error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}
	// Go to loc, follow redirects, download html,
	// parse body for <link rel="icon" href="path-to-icon">
	// and return all paths to referenced favicons
	//
	// Also: blindly try favicon.ico in the site root as fallback

	html, resp, err := f.getHTML(loc)
	if err != nil {
		log.Fatalf("cannot get HTML: %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(html)
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
			log.Println("cannot find a favicon URL in HTML body 1/2")
		}
	})
	doc.Find("head link[rel=shortcut.icon]").Each(func(i int, s *goquery.Selection) {
		href := ""
		href, found = s.Attr("href")

		icoHrefs = append(icoHrefs, href)
		if !found {
			log.Println("cannot find a favicon URL in HTML body 2/2")
		}
	})

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
			icons = append(icons, u.String())
		}
	}

	// Last resort: Try favicon.ico in document root
	if len(icons) == 0 {
		docroot := loc + "/favicon.ico"
		resp, err := client.Get(docroot)
		if err != nil {
			log.Fatalf("Cannot find %v\n", err)
			return nil, err
		}
		defer resp.Body.Close()

		icons = append(icons, docroot)
	}

	return icons, nil
}

// getFavicon downloads a favicon
func (f *Fetcher) getFavicon(url string) (icon []byte, err error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}
	if len(url) >= 7 {
		resp, err := client.Get(url)
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

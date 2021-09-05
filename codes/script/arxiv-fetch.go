/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

// go.mod >>>
// module github.com/gorun/arxiv-fetch
// go 1.13.9
// require github.com/devincarr/goarxiv latest
// require github.com/ogier/pflag latest
// require golang.org/x/tools v0.1.2
// <<< go.mod
//
// go.env >>>
// GO111MODULE=on
// <<< go.env

package main

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"io"
	"log"
	"net/http"
	"os"

	"github.com/devincarr/goarxiv"
	"github.com/ogier/pflag"
	"golang.org/x/tools/blog/atom"
)

type article struct {
	id    string
	title string
	url   string
}

func getArticles(term string, ids []string, count int) []article {
	s := goarxiv.New()
	if len(ids) == 0 {
		s.AddQuery("search_query", fmt.Sprintf("all:%v", term))
	} else {
		s.AddQuery("id_list", strings.Join(ids, ","))
	}
	s.AddQuery("max_results", strconv.Itoa(count))
	result, error := s.Get()
	if error != nil {
		fmt.Println(error)
	}

	articles := make([]article, 0)
	for _, entry := range result.Entry {
		if article, ok := getArticle(entry); ok {
			articles = append(articles, *article)
		}
	}
	log.Printf("total articles: %s", len(articles))
	return articles
}

func getArticle(e *atom.Entry) (*article, bool) {
	for _, link := range e.Link {
		if link.Type == "application/pdf" && strings.TrimSpace(e.Title) != "" && link.Href != "" {
			log.Printf("article title %s,href %s,id %v", e.Title, link.Href, e.ID)
			return &article{"", e.Title, link.Href}, true
		}
	}
	return nil, false
}

func downloadFileC(filepath string, url string, wg *sync.WaitGroup) error {
	defer wg.Done()
	return downloadFile(filepath, url)
}

func downloadFile(filepath string, url string) error {
	if url == "" {
		log.Printf("skip empty url!")
		return nil
	}
	log.Printf("downloading file %s from %s", filepath, url)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getFullName(a article, p string) string {
	filename := fmt.Sprintf("%v.pdf", strings.TrimSpace(a.title))
	spaces := regexp.MustCompile("[[:space:]]+")
	filename = spaces.ReplaceAllString(filename, " ")
	return path.Join(p, filename)
}

func downloadArticles(query string, ids []string, count int, p string, parallel bool) {
	articles := getArticles(query, ids, count)
	os.MkdirAll(p, 0755)

	if parallel {
		var wg sync.WaitGroup
		wg.Add(len(articles))
		for _, article := range articles {
			go downloadFileC(getFullName(article, p), article.url, &wg)
		}
		wg.Wait()
	} else {
		for _, article := range articles {
			downloadFile(fmt.Sprintf(getFullName(article, p), strings.TrimSpace(article.title)), article.url)
		}
	}
}

func main() {
	search := pflag.StringP("search", "s", "", "The types of articles you want to search for. 'google' by default as an example.")
	count := pflag.IntP("count", "c", 10, "The number of articles you want to retrieve. 10 by default.")
	path := pflag.StringP("path", "p", ".", "the location you want to store the articles. A folder of the current directory by default.")
	parallel := pflag.BoolP("parallel", "P", true, "Whether or not you want to pull articles down in parallel. Default is yes.")

	pflag.Parse()

	ids := pflag.Args()

	downloadArticles(*search, ids, *count, *path, *parallel)
}

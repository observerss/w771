/*

This is an exercise GO program written by observer(jingchaohu@gmail.com)

It accepts two commandline arguments namely `baseurl` and `root`
Then behave like a web crawler and save results to `root`

For more info, visit http://w771.51qiangzuo.com/

*/
package main

import (
    "fmt"
    "net/http"
    "regexp"
    "io/ioutil"
    "os"
    "encoding/base64"
)

/* Fetcher */
type Fetcher interface {
    // Fetch returns the body of URL and
    // a slice of URLs found on that page.
    Fetch(url string) (body []byte, urls []string, err error)
}

type realFetcher struct {
    baseurl string
}

func (f *realFetcher) Fetch(url string) ([]byte, []string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, nil, err
    }
    defer resp.Body.Close()
    r, _ := regexp.Compile(`href="(`+f.baseurl+`[^"]+)"`)
    bytes, _ := ioutil.ReadAll(resp.Body)
    body := string(bytes)
    urls := make([]string, 0, 10)
    links := r.FindAllStringSubmatch(body, -1)
    for _, m := range links {
        urls = append(urls, m[1])
    }
    return bytes, urls, nil
}

/* Crawler */
// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, root string) {
    type moreUrls struct{
        depth int
        urls  []string
    }
    os.Mkdir(root, 0777)
    baseurl := url
    more := make(chan moreUrls)
    
    getPage := func(url string, depth int) {
        body, urls, err := fetcher.Fetch(url)
        if err != nil {
            fmt.Printf("[err] %s\n", err.Error())
        } else {
            if url != baseurl {
                path := root + "/" + base64.StdEncoding.EncodeToString([]byte(url[len(baseurl):]))
                fmt.Printf("[found] %d, %s, %s\n", depth, url, path)
                ioutil.WriteFile(path, body, 0666)
            }
        }
        more <- moreUrls{depth+1, urls}
    }
    
    outstanding := 1
    go getPage(url, 0)
    
    visited := map[string]bool{url:true}
    for outstanding > 0 {
        next := <-more
        outstanding--
        if next.depth > depth {
            continue
        }
        for _, url := range next.urls {
            if _, seen := visited[url]; seen {
                continue
            }
            visited[url] = true
            outstanding++
            go getPage(url, next.depth)
        }
    }
}

/* Main */
func main() {
    baseurl := "http://blog.csdn.net/xushiweizh"
    root := "wushiweizh"
    
    for  i, arg := range os.Args {
        switch i { 
            case 1: baseurl = arg
            case 2: root = arg
        }

    }
    fetcher := &realFetcher{baseurl}
    Crawl(baseurl, 10, fetcher, root)
}


package searcher

import (
	"encoding/base64"
	"fmt"
	"github.com/gocolly/colly/v2"
	"hahajing/com"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Searcher interface {
	Search(q string) []*com.Ed2kFileLinkJSON
}

type ShareRipSearcher struct {
	url string
	User string
	Password string
	collyCollector colly.Collector
	results []*com.Ed2kFileLinkJSON
}
func NewShareRipSearcher(shareRipSearcher *ShareRipSearcher) *ShareRipSearcher {
	shareRipSearcher.url = "https://www.sharerip.com/forum/"
	shareRipSearcher.collyCollector = *colly.NewCollector(colly.Async(true),)
	shareRipSearcher.collyCollector.Limit(&colly.LimitRule{
		Delay:        100 * time.Millisecond,
		//RandomDelay:  100 * time.Second,
		Parallelism:  30,
	})
	shareRipSearcher.collyCollector.AllowURLRevisit=true
	shareRipSearcher.collyCollector.OnHTML("a", func(e *colly.HTMLElement) {
		//shareRipSearcher.collyCollector.AllowURLRevisit=false
		if strings.HasPrefix(e.Attr("href"),"ed2k://") {
			e2dkSplited := strings.Split(e.Attr("href"),"|")
			name,_:= url.QueryUnescape(e2dkSplited[2])
			size,_:=strconv.ParseInt(e2dkSplited[3],10,64)
			hash:=e2dkSplited[4]
			link := fmt.Sprintf("%s|%d|%s" ,name,size,hash)
			bytes:=[]byte(link)
			downloadLink :="/download?file="+base64.URLEncoding.EncodeToString(bytes)
			e2dkFile := com.Ed2kFileLinkJSON{
				FileInfo:     com.FileInfo{
					Type:    1,
					OrgName: "latest",
					ChName:  "",
					Season:  0,
					Episode: 0,
				},
				Name:         name,
				Season:       0,
				Episode:      0,
				Size:         uint64(size),
				Avail:        10,
				Link:         e.Attr("href"),
				DownloadLink: downloadLink,
			}
			shareRipSearcher.results = append(shareRipSearcher.results,&e2dkFile)
		}
		//else if strings.HasPrefix(e.Attr("href"),"https://www.sharerip.com") && !strings.Contains(e.Attr("href"),"logout")   {
		//	match,_ := regexp.MatchString(".*/forum/index.php\\?(action=forum|board|topic).*", e.Attr("href"))
		//	if match {
		//			shareRipSearcher.collyCollector.Visit(e.Attr("href"))
		//	}
		//}
	})

	shareRipSearcher.collyCollector.OnHTML("form#frmLogin", func(e *colly.HTMLElement) {
		postURL := e.Attr("action")
		var tokenElement *colly.HTMLElement
		e.ForEach("input[type=hidden]", func(_ int, elem *colly.HTMLElement) {
			if elem.Attr("value")!=""{
				tokenElement = elem
			}
		})
		hashPassword := com.SHA1String(com.SHA1String(shareRipSearcher.User + shareRipSearcher.Password) + tokenElement.Attr("value"))
		err := shareRipSearcher.collyCollector.Post(postURL, map[string]string{
			"user": shareRipSearcher.User,
			"passwrd": shareRipSearcher.Password,
			"cookieneverexp":"on",
			tokenElement.Attr("name"):tokenElement.Attr("value"),
			"hash_passwrd":hashPassword})
		if err != nil {
			com.HhjLog.Error("Error on login on sharerip: ", err)
		}
	})

	shareRipSearcher.collyCollector.OnRequest(func(r *colly.Request) {
		log.Println("request", r.URL)
	})
	return shareRipSearcher
}

func (srs *ShareRipSearcher) Search(q string) []*com.Ed2kFileLinkJSON {
	srs.results = nil
	srs.collyCollector.Visit(srs.url)
	srs.collyCollector.Wait()
	return srs.results
}
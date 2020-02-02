package web

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gorilla/mux"
	"hahajing/com"
	"hahajing/kad"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

const (
	keywordCheckWaitingTime = 5
	kadSearchWaitingTime    = 15
)

// webError is for user browser.
type webError struct {
	Error string
}

// Web x
type Web struct {
	searchReqCh       chan *kad.SearchReq





}

// Start x
func (we *Web) Start(searchReqCh chan *kad.SearchReq, ) {
	we.searchReqCh = searchReqCh
	we.startServer()
}



func (we *Web) readSearchInput(query string) (*com.MyKeyword) {
	keywords := com.Split2Keywords(query)
	myKeyword := com.NewMyKeyword(keywords)
	return myKeyword
}


func (we *Web) writeError(ws *websocket.Conn, errStr string) {
	data, _ := json.Marshal(&webError{Error: errStr})
	ws.Write(data)
}

func (we *Web) send2Kad(myKeywordStruct *com.MyKeywordStruct) []*com.Ed2kFileLinkJSON  {
	resCh := make(chan *kad.SearchRes, kad.SearchResChSize)
	timeoutChannel := make(chan bool, 1)
	lastDateFound :=  time.Now().Unix()+ kad.SearchTimeWithoutResults

	var results []*com.Ed2kFileLinkJSON
	searchReq := kad.SearchReq{ResCh: resCh, MyKeywordStruct: myKeywordStruct}
	we.searchReqCh <- &searchReq
	for {
		select {
		default:
			now := time.Now().Unix()
			if now > lastDateFound {
				timeoutChannel <- true
			}
			time.Sleep(100*time.Millisecond)
		case pSearchRes := <-resCh:
			for _, fileLink := range pSearchRes.FileLinks {
				contains:=true
				for  _,targetWord := range myKeywordStruct.MyKeyword.SearchKeywords {
					if !strings.Contains(strings.ToLower(fileLink.Name),targetWord) {
						contains=false
					}
				}
				if contains {
					//com.HhjLog.Infof("Elements found %d", len(pSearchRes.FileLinks))
					results = append(results,fileLink.ToJSON())
				}else{
					com.HhjLog.Infof("Not filtered %s",fileLink.Name)
				}
			}
			lastDateFound = time.Now().Unix()+ kad.SearchTimeWithoutResults
			if pSearchRes.Cached {
				timeoutChannel <- true
			}

		case <- timeoutChannel:
			com.HhjLog.Infof("Elements found %d", len(results))
			return results
		}
	}
}
func (we *Web) downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileURL:=r.URL.Query().Get("file")
	//fileURL = fileURL + "="
	bytesLink,err :=base64.URLEncoding.DecodeString(fileURL)
	if err !=nil {
		fmt.Printf("Error decoding string: %s ", err.Error())
	}
	bytesReader := bytes.NewReader(bytesLink)
	downloadLinkDecodedSplited := strings.Split(fmt.Sprintf("%s", bytesLink),"|")
	name:=downloadLinkDecodedSplited[0]
	size,_:=strconv.Atoi(downloadLinkDecodedSplited[1])
	hash:=downloadLinkDecodedSplited[2]
	if len(hash) != 32 {
		println("Invalid Hash:" + hash + "|" + name)
	}

	metaInfo:= metainfo.MetaInfo{
		Announce:     fmt.Sprintf("ed2k://|file|%s|%d|%s|/",com.EncodeURLUtf8(com.StripInvalidFileNameChars(name)),size,hash),
		CreationDate: 0,
		Comment:      "Fake torrent E2dk",
		CreatedBy:    "nobody",
	}
	pieceLength :=  int64(1048576)
	//jsonReader, _ := io.CopyN(w, bytesReader, int64(len(bytesLink)))
	var pieces []byte
	for {
		hasher := sha1.New()
		wn, err := io.CopyN(hasher, bytesReader, int64(len(bytesLink)))
		if err == io.EOF {
			err = nil
		}
		if wn == 0 {
			break
		}
		pieces = hasher.Sum(pieces)
		if wn < pieceLength {
			break
		}
	}
	nameCut := 70
	if len(name) < 70 {
		nameCut = len(name)
	}
	info := metainfo.Info{
		Name: name[0:nameCut],
		PieceLength: pieceLength,
		Length: int64(len(bytesLink)),
		Pieces:      pieces,
		Files: []metainfo.FileInfo{
			{Path: []string{name}, Length: int64(len(bytesLink))},
		},
	}
	metaInfo.InfoBytes,_ = bencode.Marshal(info)
	w.Header().Add("Content-Disposition",fmt.Sprintf("attachment; filename=%s.torrent",name))
	metaInfo.Write(w)
}

func (we *Web) fakeQBittorrent(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.RequestURI(),"login"){
		w.WriteHeader(200)
		w.Header().Add("Content-Length","3")
		w.Write([]byte("Ok."))
	}else if strings.Contains(r.URL.RequestURI(),"webapiVersion"){
		w.WriteHeader(200)
		w.Header().Add("Content-Length","3")
		w.Write([]byte("2.4"))
	}else if strings.Contains(r.URL.RequestURI(),"torrents/add"){
		//r.MultipartForm.File
		r.ParseForm()
		file,_, err := r.FormFile("torrents")
		if err == nil {
			defer file.Close()
			//buf := new(bytes.Buffer)
			//buf.ReadFrom(file)

			metaInfo, err := metainfo.Load(file)
			if err == nil {
				client := &http.Client{}
				form := url.Values{}
				form.Add("p", "cambiarPassword")
				form.Add("w", "password")
				req, err := http.NewRequest("POST", "http://localhost:4711", strings.NewReader(form.Encode()))
				if err == nil {
					response, err := client.Do(req)
					if err==nil {
						if response.StatusCode == 200 {
							buf := new(bytes.Buffer)
							buf.ReadFrom(response.Body)
							content := buf.String()
							pattern := regexp.MustCompile(`ses=(-?\d*)&`)
							ses:= pattern.FindStringSubmatch(content)[1]
							uploadURL := "http://localhost:4711/?ses=" + ses + "&w=transfer&ed2k="+url.QueryEscape(metaInfo.Announce)
							response,err := http.Get(uploadURL)
							if err == nil {
								println(response.Status)
							}
						}else{
							err = errors.New("invalid status:"+ response.Status)
						}
					}
				}
			}
			if err == nil {
				w.WriteHeader(200)
				w.Header().Add("Content-Length","3")
				w.Write([]byte("Ok."))
			}else{
				w.WriteHeader(500)
				w.Header().Add("Content-Length","3")
				w.Write([]byte("KO."))
			}

		}else{
			w.WriteHeader(404)
		}

	}else{
		println(r.URL.RequestURI())
	}
}

func (we *Web) searchHandler(w http.ResponseWriter, r *http.Request) {
	q:=r.URL.Query().Get("q")
	myKeyword := we.readSearchInput(q)

	// send to KAD
    var items []*com.Item
    item := com.Item{Type: 0x1, OrgName: strings.Join(myKeyword.SearchKeywords, " "), ChName: ""}
    items = append(items, &item)
    com.HhjLog.Infof("New search: %#v", items[0])
	myKeywordStruct := com.NewMyKeywordStruct(myKeyword, items)
	results :=we.send2Kad(myKeywordStruct)

	bytes,_ :=xml.MarshalIndent(results,"","   ")
	//bytes,_ :=json.MarshalIndent(results,"","\t")
	w.Write(bytes)
}




func (we *Web) startServer() {
	com.HhjLog.Info("Web Server is running...")
	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/{name:.*}", we.fakeQBittorrent)
	http.Handle("/api/", rtr)
	http.HandleFunc("/search", we.searchHandler)
	http.HandleFunc("/download", we.downloadHandler)

	var err error
	if len(os.Args) > 1 && os.Args[1] == "server" {
		err = http.ListenAndServe(":80", nil)
	} else {
		err = http.ListenAndServe(":66", nil)
	}
	if err != nil {
		com.HhjLog.Panic("Start Web Server failed: ", err)
	}
}

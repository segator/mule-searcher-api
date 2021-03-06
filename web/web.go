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
	"hahajing/download"
	"hahajing/kad"
	"hahajing/searcher"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)


// webError is for user browser.
type webError struct {
	Error string
}

// Web x
type Web struct {
	searchReqCh chan *kad.SearchReq
	config *com.Config
	e2dkDownloader download.E2DKDownloader
	torrentDownloader download.TorrentDownloader
	searcher searcher.Searcher
}
type qBittorrentPreferences struct{
	SavePath string `json:"save_path"`
	MaxRatioEnabled bool `json:"max_ratio_enabled"`
	MaxRatio float32 `json:"max_ratio"`
	MaxSeedingTimeEnabled bool `json:"max_seeding_time_enabled"`
	MaxSeedingTime int32 `json:"max_seeding_time"`
	MaxRatioACT bool `json:"max_ratio_act"`
	QueueingEnabled bool `json:"queueing_enabled"`
	Dht bool `json:"dht"`
}

// Start x
func (we *Web) Start(searchReqCh chan *kad.SearchReq,config *com.Config, e2dkDownloader download.E2DKDownloader,torrentDownloader download.TorrentDownloader,searcher searcher.Searcher) {
	we.searchReqCh = searchReqCh
	we.config = config
	we.e2dkDownloader = e2dkDownloader
	we.torrentDownloader = torrentDownloader
	we.searcher = searcher
	go we.startServer()
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
	lastDateFound :=  time.Now().Unix()+ int64(we.config.SearchTimeWithoutResults)

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
					com.HhjLog.Debugf("Elements found %d", len(pSearchRes.FileLinks))
					results = append(results,fileLink.ToJSON())
				}else{
					com.HhjLog.Infof("Not filtered %s",fileLink.Name)
				}
			}
			lastDateFound = time.Now().Unix()+ int64(we.config.SearchTimeWithoutResults)
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
		r.ParseForm()
		if r.FormValue("password") == we.config.HTTPPassword {
			cookie := http.Cookie{
				Name:    "token",
				Value:   r.FormValue("password"),
				SameSite: http.SameSiteStrictMode,
				Path:"/",


			}
			http.SetCookie(w, &cookie)
			w.WriteHeader(200)
			w.Header().Add("Content-Length","3")
			w.Write([]byte("Ok."))

		}else{
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
	}else if strings.Contains(r.URL.RequestURI(),"webapiVersion") {
		w.WriteHeader(200)
		w.Header().Add("Content-Length", "3")
		w.Write([]byte("2.4"))
	}else if strings.Contains(r.URL.RequestURI(),"app/preferences") {
		w.WriteHeader(200)
		preferences := &qBittorrentPreferences{
			SavePath:              we.config.DownloadPath,
			MaxRatioEnabled:       false,
			MaxRatio:              0,
			MaxSeedingTimeEnabled: false,
			MaxSeedingTime:        0,
			MaxRatioACT:           false,
			QueueingEnabled:       false,
			Dht:                   false,
		}
		bytes,_ :=json.MarshalIndent(preferences,"","   ")
		w.Header().Add("Content-Length",strconv.Itoa(len(bytes)))
		w.Write(bytes)
	}else if strings.Contains(r.URL.RequestURI(),"torrents/add"){
		//r.MultipartForm.File
		token:=r.URL.Query().Get("token")
		if (len(r.Cookies()) > 0  && r.Cookies()[0].Name == "token" && r.Cookies()[0].Value ==  we.config.HTTPPassword) || token == we.config.HTTPPassword {
			r.ParseForm()
			file,_, err := r.FormFile("torrents")
			if err == nil {
				defer file.Close()
				//buf := new(bytes.Buffer)
				//buf.ReadFrom(file)

				buf, _ := ioutil.ReadAll(file)
				buf1 := bytes.NewBuffer(buf)
				buf2 := bytes.NewBuffer(buf)
				metaInfoFile, err := metainfo.Load(buf1)
				if err == nil {
					info,err := metaInfoFile.UnmarshalInfo()

					if err == nil {
						var downloadResult bool
						com.HhjLog.Infof("Downloading: %s",info.Name)
						if strings.HasPrefix(metaInfoFile.Announce,"ed2k://"){
							downloadResult = we.e2dkDownloader.DownloadE2DK(metaInfoFile.Announce)
						}else{
							bytes, _ := ioutil.ReadAll(buf2)
							downloadResult = we.torrentDownloader.DownloadTorrent(info.Name,bytes)
						}
						if !downloadResult {
							err = errors.New("Failed to download" + info.Name)
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
			w.WriteHeader(403)
		}
	}else{
		fmt.Println(r.URL.RequestURI())
	}
}

func (we *Web) searchHandler(w http.ResponseWriter, r *http.Request) {
	var results []*com.Ed2kFileLinkJSON
	q:=r.URL.Query().Get("q")
	t:=r.URL.Query().Get("t")
	var typeSearch byte
	switch t {
		case "tvshow":
			typeSearch = com.SeasonTV
		case "movie":
			typeSearch = com.Movie
		default:
			typeSearch = com.UnknownType
	}

	if q!="" {
		myKeyword := we.readSearchInput(q)
		var items []*com.Item
		item := com.Item{Type: typeSearch, OrgName: strings.Join(myKeyword.SearchKeywords, " "), ChName: ""}
		items = append(items, &item)
		myKeywordStruct := com.NewMyKeywordStruct(myKeyword, items)
		results =we.send2Kad(myKeywordStruct)
	}else{
		//For now we only use searchers for getting latests
		com.HhjLog.Infof("getting latest updates from searcher")
		results = we.searcher.Search("",typeSearch)
	}
	bytes,_ :=xml.MarshalIndent(results,"","   ")
	w.Write(bytes)
}

func BasicAuth(handler http.HandlerFunc,  token string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		t:=r.URL.Query().Get("token")

		if t != token {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}



func (we *Web) startServer() {
	com.HhjLog.Info("Web Server is running...")
	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/{name:.*}", we.fakeQBittorrent)
	http.Handle("/api/", rtr)
	http.HandleFunc("/search", BasicAuth(we.searchHandler,we.config.HTTPPassword))
	http.HandleFunc("/download", we.downloadHandler)

	err := http.ListenAndServe(":"+strconv.Itoa(we.config.WEBListenPort), nil)
	if err != nil {
		com.HhjLog.Panic("Start Web Server failed: ", err)
	}
}

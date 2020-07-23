package download

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hahajing/com"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type E2DKDownloader interface {
	DownloadE2DK(e2dk string) bool

}
type TorrentDownloader interface {
	DownloadTorrent(fileName string,torrentData []byte) bool
}

type MultiDownloader struct {
	DownloaderE2dkList [] E2DKDownloader
	DownloaderTorrentList [] TorrentDownloader
}

type EmuleDownloader struct {
	EmuleWebURL string
	Password string
}
type SynologyMuleDownloader struct {
	SynologyURL string
	SynologyUser string
	SynologyPassword string
	SynologyDestionation string
}
type AmuleDownloader struct {
	AmuleHost string
	AmulePort int
	AmulePassword string
}
type SynologyAuthResponse struct {
	Data struct {
		Sid string `json:"sid"`
	} `json:"data"`
	Success bool `json:"success"`
}
func parseParams(url *url.URL,params map[string]string)  {
	query := url.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	url.RawQuery = strings.Replace(query.Encode(), "+", "%20", -1)
}
func (md MultiDownloader) DownloadTorrent(fileName string,torrentData []byte) bool {
	var failedDownloaders []int
	downloadSuccess:=false
	fail := false

	for randomNumber := 0; !downloadSuccess && !fail; randomNumber = rand.Intn(len(md.DownloaderTorrentList)) {
		alreadyFailed := false
		for _, failedIndex := range failedDownloaders{
			if failedIndex==randomNumber {
				alreadyFailed=true
				continue
			}
		}
		if alreadyFailed{
			continue
		}
		downloadSuccess = md.DownloaderTorrentList[randomNumber].DownloadTorrent(fileName,torrentData)
		if !downloadSuccess{
			failedDownloaders = append(failedDownloaders,randomNumber)
			if len(failedDownloaders) == len(md.DownloaderTorrentList) {
				fail=true
			}
		}
	}
	return downloadSuccess
}
func (md MultiDownloader) DownloadE2DK(e2dk string) bool {
	var failedDownloaders []int
	downloadSuccess:=false
	fail := false

	for randomNumber := 0; !downloadSuccess && !fail; randomNumber = rand.Intn(len(md.DownloaderE2dkList)) {
		alreadyFailed := false
		for _, failedIndex := range failedDownloaders{
			if failedIndex==randomNumber {
				alreadyFailed=true
				continue
			}
		}
		if alreadyFailed{
			continue
		}
		downloadSuccess = md.DownloaderE2dkList[randomNumber].DownloadE2DK(e2dk)
		if !downloadSuccess{
			failedDownloaders = append(failedDownloaders,randomNumber)
			if len(failedDownloaders) == len(md.DownloaderE2dkList) {
				fail=true
			}
		}
	}
	return downloadSuccess
}

func (ad AmuleDownloader) DownloadE2DK(e2dk string) bool {
	var outbuf, errbuf bytes.Buffer
	binary, err := exec.LookPath("amulecmd")
	if err != nil {
		ex, err := os.Executable()
		if err != nil {
			return false
		}
		exPath := filepath.Dir(ex)
		binary = exPath + "/amulecmd"
	}
	args := []string{"--host="+ad.AmuleHost,"--port="+strconv.Itoa(ad.AmulePort), "--password="+ad.AmulePassword, "--command=add "+e2dk+""}
	com.HhjLog.Infof("%s %v",binary,args)
	cmd := exec.Command(binary, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err = cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0  {
				com.HhjLog.Errorf("Error on execute %s %v exit code %d",binary,args,exitError.ExitCode())
				return false
			}
		}else{
			com.HhjLog.Errorf("Error on execute %s %v",binary,args)
			return false
		}
	}
	return true
}

func (ed SynologyMuleDownloader) Login(httpClient *http.Client) (string,bool) {
	url,_ := url.Parse(ed.SynologyURL + "/webapi/auth.cgi")
	loginParams :=map[string]string{
		"api":     "SYNO.API.Auth",
		"version": "2",
		"method":  "login",
		"account": ed.SynologyUser,
		"passwd":  ed.SynologyPassword,
		"session": "DownloadStation",
		"format":  "sid",
	}
	parseParams(url,loginParams)
	resp, err := httpClient.Get(url.String())
	defer resp.Body.Close()
	if err != nil || resp.StatusCode!=200 {
		err = errors.New("Failed to Login to Synology:" + ed.SynologyURL)
		return "",false
	}else {
		response := new(SynologyAuthResponse)
		json.NewDecoder(resp.Body).Decode(response)
		if !response.Success {
			return "",false
		}else{
			return response.Data.Sid,true
		}
	}
}

func (ed SynologyMuleDownloader) DownloadTorrent(fileName string,torrentData []byte) bool {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}
	sid,success := ed.Login(httpClient)
	if success {
		urlSynology,_ := url.Parse(ed.SynologyURL + "/webapi/DownloadStation/task.cgi")
		downloadParams :=map[string]string{
			"api":     "SYNO.DownloadStation.Task",
			"version": "1",
			"method":  "create",
			"_sid":sid,
			"destination":ed.SynologyDestionation,
		}

		payload := &bytes.Buffer{}
		w := multipart.NewWriter(payload)
		parseFormData(downloadParams,w)
		if fw, err := w.CreateFormFile("file",fileName + ".torrent"); err != nil {
			return false
		}else{
			int,err := fw.Write(torrentData)
			if err!=nil {
				fmt.Println(err)
			}
			fmt.Println(int)
		}
		w.Close()


		req, err := http.NewRequest("POST", urlSynology.String(), payload)
		req.Header.Set("Content-Type", w.FormDataContentType())

		if err != nil {
			return false
		}
		resp, err := httpClient.Do(req)
		if  err!= nil || resp.StatusCode !=200{
			err = errors.New("Failed to Download torrent:" + fileName)
		}
		defer resp.Body.Close()
		return err == nil
	}else{
		return false
	}
}

func parseFormData(params map[string]string, w *multipart.Writer) {
	for k, v := range params {
		_ = w.WriteField(k,v)
	}
}



func (ed SynologyMuleDownloader) DownloadE2DK(e2dk string) bool {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}
	sid,success := ed.Login(httpClient)
	if success {
		url,_ := url.Parse(ed.SynologyURL + "/webapi/DownloadStation/task.cgi")
		downloadParams :=map[string]string{
			"api":     "SYNO.DownloadStation.Task",
			"version": "1",
			"method":  "create",
			"uri": e2dk,
			"_sid":sid,
			"destionation": ed.SynologyDestionation,
		}
		parseParams(url,downloadParams)
		resp, err := httpClient.Get(url.String())
		if  err!= nil || resp.StatusCode !=200{
			err = errors.New("Failed to Download:" + e2dk)
		}
		defer resp.Body.Close()
		return err == nil
	}else{
		return false
	}
}
func (ed EmuleDownloader) DownloadE2DK(e2dk string) bool {
	client := &http.Client{}
	form := url.Values{}
	form.Add("p", ed.Password)
	form.Add("w", "password")
	req, err := http.NewRequest("POST", ed.EmuleWebURL, strings.NewReader(form.Encode()))
	if err !=nil {
		return false
	}else  {
		response, err := client.Do(req)
		if err!=nil || response.StatusCode != 200 {
			return false
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		content := buf.String()
		pattern := regexp.MustCompile(`ses=(-?\d*)&`)
		ses:= pattern.FindStringSubmatch(content)[1]
		uploadURL := ed.EmuleWebURL +"/?ses=" + ses + "&w=transfer&ed2k="+url.QueryEscape(e2dk)
		response,err = http.Get(uploadURL)
		if err != nil || response.StatusCode!=200 {
			return false
		}

	}
	return true
}
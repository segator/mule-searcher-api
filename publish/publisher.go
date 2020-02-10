package publish

import (
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"
	"hahajing/com"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)
const (
	publisherTimer               = 120

)

type Publisher interface {
	Start() bool
}

type PublisherConfig struct {
	DownloadPath             string
	ValidUploadableFormats []string
}

type PublisherSSHConfig struct {
	Config PublisherConfig
	PublishSSHHost           string
	PublishSSHUsername		 string
	PublishSSHPassword       string
	PublishSSHPath           string
	PublishSSHPort           int
	sshClientConfig          ssh.ClientConfig
	sshClient                scp.Client
}

type FileInfo struct {
	path string
	info os.FileInfo

}

func (p *PublisherSSHConfig) Start() bool {
	if p.PublishSSHHost != "" {
		if  p.PublishSSHPassword != "" {
			p.sshClientConfig,_ = auth.PasswordKey(p.PublishSSHUsername,p.PublishSSHPassword, ssh.InsecureIgnoreHostKey())
		}else{
			com.HhjLog.Error("No SSH Password were set")
			return false
		}
		p.sshClient = scp.NewClient(p.PublishSSHHost+":"+strconv.Itoa(p.PublishSSHPort), &p.sshClientConfig)
		go p.scheduleRoutine()
	}
	return true
}

func (p *PublisherSSHConfig) scheduleRoutine() {
	tick := time.NewTicker(publisherTimer * time.Second)
	for {
		select {
		case <-tick.C:
			uploadableFiles,err := getUploadableFiles(p.Config.DownloadPath,p.Config.ValidUploadableFormats)
			if err!=nil {
				com.HhjLog.Error("Error when processing download folder: %s %s", p.Config.DownloadPath, err)
			}else{
				for _,uploadableFile := range uploadableFiles {
					err := p.sshClient.Connect()
					if err != nil {
						com.HhjLog.Error("Couldn't establish a connection to the remote server ", err)
						continue
					}

					f, _ := os.Open(uploadableFile.path+"/"+uploadableFile.info.Name())
					fileInfo, _ := f.Stat()
					p.sshClient.Connect()
					out,err := p.sshClient.Session.Output("ls -lah \"" + p.PublishSSHPath+"/"+uploadableFile.info.Name()+ "\" > /dev/null 2>&1  && echo $?")
					p.sshClient.Close()
					if string(out) != "0\n"{
						p.sshClient.Connect()
						com.HhjLog.Info("Uploading " + uploadableFile.path+"/"+uploadableFile.info.Name() + " on " + p.PublishSSHHost + ":" +strconv.Itoa(p.PublishSSHPort)+" "+p.PublishSSHPath)
						p.sshClient.Copy(f, p.PublishSSHPath+"/"+uploadableFile.info.Name(), "0655",fileInfo.Size())
						p.sshClient.Close()
					}
					f.Close()
					if err != nil {
						com.HhjLog.Error("Error on publish " + uploadableFile.path +"/"+uploadableFile.info.Name() + "  on" + p.PublishSSHHost + ":" +strconv.Itoa(p.PublishSSHPort)+" "+p.PublishSSHPath)
					}

				}
			}
		}
	}
}

func getUploadableFiles(path string, extensions []string) ([]FileInfo,error)  {
	var files []FileInfo
	nowTime := time.Now()
	ReadedFiles,err :=ioutil.ReadDir(path)
	if err!= nil {
		return nil,err
	}
	for _,info := range ReadedFiles {
		toLower := strings.ToLower(info.Name())
		uploadablePeriodTime := info.ModTime().Add(5 * time.Minute)
		if !info.IsDir() && uploadablePeriodTime.Before(nowTime) {
			for _,extension := range extensions {
				if strings.HasSuffix(toLower,extension) {
					files = append(files, FileInfo{
						path: path,
						info: info,
					})
				}
			}
		}
	}
	return files,err
}

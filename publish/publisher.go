package publish

import (
	"errors"
	"fmt"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"hahajing/com"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)
const (
	publisherTimer               = 1

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
					err := p.uploadFile(uploadableFile)
					if err !=nil {
						com.HhjLog.Errorf("An error ocurrend when uploading %s",uploadableFile.path,err)
					}

				}
			}
		}
	}
}

func (p *PublisherSSHConfig) uploadFile(sourceFile FileInfo) error {
	conn, err := ssh.Dial("tcp", p.PublishSSHHost+":"+strconv.Itoa(p.PublishSSHPort), &p.sshClientConfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	f, _ := os.Open(sourceFile.path+"/"+sourceFile.info.Name())
	defer f.Close()
	sourceFileStat,_:=f.Stat()
	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()
	destFileString := p.PublishSSHPath+"/"+sourceFile.info.Name()
	destFileStat,_ := client.Stat(destFileString)

	if destFileStat == nil || destFileStat.Size() != sourceFileStat.Size() {
		dstFile,err := client.Create(destFileString)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		com.HhjLog.Info("Uploading " + sourceFile.path+"/"+sourceFile.info.Name() + " ---> " + p.PublishSSHHost + ":" +strconv.Itoa(p.PublishSSHPort)+""+p.PublishSSHPath)
		bytes, err := io.Copy(dstFile, f)
		if err != nil {
			return err
		}


		if bytes != sourceFileStat.Size() {
			return  errors.New(fmt.Sprintf("%d bytes copied from %d ", bytes, sourceFileStat.Size()))
		}
	}
	return err
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
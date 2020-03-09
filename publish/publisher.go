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


type Publisher interface {
	Start() bool
}

type PublisherConfig struct {
	DownloadPath             string
	DownloadPathCompleted    string
	ValidUploadableFormats []string
}

type PublisherSSHConfig struct {
	Config PublisherConfig
	ScanTime                 time.Duration
	PublishSSHHost           string
	PublishSSHUsername		 string
	PublishSSHPassword       string
	PublishSSHPathTV         string
	PublishSSHPathMovies     string
	PublishSSHPort           int
	PublishMinimumTime       time.Duration
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
	tick := time.NewTicker(p.ScanTime)
	for {
		select {
		case <-tick.C:
			uploadableFiles,err := getUploadableFiles(p.Config.DownloadPath,p.Config.ValidUploadableFormats,p.PublishMinimumTime)
			if err!=nil {
				com.HhjLog.Errorf("Error when processing download folder: %s %s", p.Config.DownloadPath, err)
			}else{
				for _,uploadableFile := range uploadableFiles {
					err := p.uploadFile(uploadableFile)
					if err !=nil {
						com.HhjLog.Errorf("An error ocurrend when uploading %s %s",uploadableFile.path,err)
					} else if p.Config.DownloadPathCompleted!="" {
						destinationMovePath := p.Config.DownloadPathCompleted + "/" + uploadableFile.info.Name()
						err := os.Rename(uploadableFile.path + "/" + uploadableFile.info.Name(), destinationMovePath)
						if err != nil {
							com.HhjLog.Errorf("An error ocurrend when moving %s to %s %s",uploadableFile.path, destinationMovePath,err)
						}
					}

				}
			}
		}
	}
}

func (p *PublisherSSHConfig) uploadFile(sourceFile FileInfo) error {
	_,_,videoType := com.ParseUnknownTypeName(sourceFile.info.Name(),".*")
	var sshPath string
	switch videoType {
		case com.Movie:
			sshPath = p.PublishSSHPathMovies
		case com.SeasonTV:
			sshPath = p.PublishSSHPathTV
		case com.NoSeasonTV:
			sshPath = p.PublishSSHPathTV
	}

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
	destFileString := sshPath+"/"+sourceFile.info.Name()
	destFileStat,_ := client.Stat(destFileString)

	if destFileStat == nil || destFileStat.Size() != sourceFileStat.Size() {
		dstFile,err := client.Create(destFileString)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		com.HhjLog.Infof("Uploading name:%s/%s size: %dMB ---> %s:%d/%s",sourceFile.path,sourceFile.info.Name(),sourceFileStat.Size()/1048576,p.PublishSSHHost,p.PublishSSHPort,sshPath)
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

func getUploadableFiles(path string, extensions []string,minimumTime time.Duration) ([]FileInfo,error)  {
	var files []FileInfo
	nowTime := time.Now()
	ReadedFiles,err :=ioutil.ReadDir(path)
	if err!= nil {
		return nil,err
	}
	for _,info := range ReadedFiles {
		toLower := strings.ToLower(info.Name())
		uploadablePeriodTime := info.ModTime().Add(minimumTime)
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
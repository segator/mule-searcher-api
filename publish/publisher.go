package publish

import (
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"hahajing/com"
	"io"
	"io/ioutil"
	"log"
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
	sshConnection            ssh.Client
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
		conn, err := ssh.Dial("tcp", p.PublishSSHHost+":"+strconv.Itoa(p.PublishSSHPort), &p.sshClientConfig)
		if err != nil {
			log.Fatal(err)
		}
		p.sshConnection = *conn
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
					f, _ := os.Open(uploadableFile.path+"/"+uploadableFile.info.Name())

					session, err := p.sshConnection.NewSession()
					if err != nil {
						com.HhjLog.Error("Error when openning ssh session ", err)
						continue
					}
					out,err := session.Output("ls -lah \"" + p.PublishSSHPath+"/"+uploadableFile.info.Name()+ "\" > /dev/null 2>&1  && echo $?")
					defer session.Close()
					if string(out) != "0\n"{
						client, err := sftp.NewClient(&p.sshConnection)
						if err != nil {
							com.HhjLog.Error("Couldn't establish a connection to the remote server ", err)
							continue
						}
						destFileString := p.PublishSSHPath+"/"+uploadableFile.info.Name()
						dstFile,err := client.Create(destFileString)
						if err != nil {
							com.HhjLog.Error("Error on create destination file ",destFileString,err)
							continue
						}

						com.HhjLog.Info("Uploading " + uploadableFile.path+"/"+uploadableFile.info.Name() + " ---> " + p.PublishSSHHost + ":" +strconv.Itoa(p.PublishSSHPort)+""+p.PublishSSHPath)
						bytes, err := io.Copy(dstFile, f)
						if err != nil {
							log.Fatal(err)
						}
						dstFile.Close()
						f.Close()
						com.HhjLog.Infof("%d bytes copied\n", bytes)
					}

					if err != nil {
						exitError,ok := err.(*ssh.ExitError)
						if ok{
							com.HhjLog.Error("SSH Error on publish " + strconv.Itoa(exitError.ExitStatus()))
						}else{
							com.HhjLog.Error("Error on publish " + uploadableFile.path +"/"+uploadableFile.info.Name() + "  on " + p.PublishSSHHost + ":" +strconv.Itoa(p.PublishSSHPort)+" "+p.PublishSSHPath)
						}
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

# kad-e2dk-searcher
This application allow to find e2dk on KAD Network via API, thanks to https://github.com/moyuanz/hahajing
I created this application to allow sickrage and forks to support downloading of e2dk.

![Build](https://github.com/segator/mule-searcher-api/workflows/Build/badge.svg)

This application works with jackett, check my fork of jackett to support mule-searcher-api
https://github.com/segator/Jackett/tree/feature/e2dkkad
and is tested with Medusa(Sickrage fork)

1. **Sickrage** asks to jackett to search new e2dk
2. **Jackett** Connect receive de search request and send it to kad-e2dk-searcher
3. **kad-e2dk-searcher** Search on KAD network and returns back to jackett
4. **Jackett** return e2dk list to sickrage
5. **Sickrage** Decide to download a new **Torrent** **Wait you said torrent? we wanted to download e2dk!!** *Sickrage doesn't know how to procees e2dk links so I faked the system creating fake torrents with the e2dk info inside.*
6. **Sickrage** Send the torrent to a fake **Qbittorrent** implementation in kad-e2dk-searcher
7. **kad-e2dk-searcher** extract the e2dk of the torrent and send it to your configured downloader(amulecmd,synology,emuleweb) 
8. *(Optional)* **kad-e2dk-searcher** After file is downloaded send the files via SFTP to your desired server.  

## Build
```
go get
go build
```

## RUN
Can be done using the binary or the docker image 

### Command Line
```bash
  -NumberOfUDPSockets int
    	Number of UDP Sockets (default 10)
  -contacts int
    	Max number of contacts (default 5000)
  -emule-password string
    	admin (default "admin")
  -emule-url string
    	Emule URL (default "http://localhost:4711")
  -nodesdat-path string
    	nodes.dat path can be http:// or file:// (default "http://upd.emule-security.org/nodes.dat")
  -public-udp-listen-port int
    	Public UDP Listen Port (default 2500)
  -search-cache-timeout int
    	Time to cache searches (default 60)
  -tcp-listen-port int
    	TCP Listen Port (default 2501)
  -timeout-noresults int
    	Time to finish search after no more results (default 8)
  -udp-listen-port int
    	UDP Listen Port (default 2500)
  -web-listen-port int
    	Web Listen Port (default 80)
```

### Docker
```
docker run -d --name emule-searcher --restart=always -p 35115:35115/udp -p 8080:80 -v /volume2/downloads/download/:/downloads \
            segator/kad-e2dk-api -http-password admin -udp-listen-port 35115 -public-udp-listen-port 35115
```
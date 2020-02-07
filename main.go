package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kelseyhightower/confd/log"
	"io"
	"net/http"
	"os"
	"strconv"
)

var ablumPath string
func main() {
	albumIdArg := flag.String("id", "", "albumId")
	dirArg:= flag.String("dir", "", "download file path")
	flag.Parse()

	albumId := *albumIdArg
	dir := *dirArg

	if albumId == "" {
		log.Info("error albumId")
		return
	}
	if dir == "" {
		dir = "/Users/springcat/Downloads"
	}

	title := getAblumTitle(albumId)
	ablumPath = dir + "/" + title
	err := os.MkdirAll(ablumPath, os.ModePerm)
	assertOk(err)
	log.Info("mkdir ablumPath:"+ablumPath)
	log.Info("start download album")
	handleAlbumPage(albumId,1)
	log.Info("success download ablum %s",title)
}

func getAblumTitle(albumId string) string {
	resp, err := http.Get("https://www.ximalaya.com/revision/album?albumId="+albumId)
	assertOk(err)
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var album Album
	err = decoder.Decode(&album)
	assertOk(err)
	return album.Data.MainInfo.AlbumTitle
}

func handleAlbumPage(albumId string,pageNum int)  {
	log.Info("download pageNum:%d",pageNum)
	url := fmt.Sprintf("https://www.ximalaya.com/revision/album/v1/getTracksList?albumId=%s"+"&pageNum=%d", albumId, pageNum)
	resp, err := http.Get(url)
	assertOk(err)
	//parse json
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var tracksList TracksList
	err = decoder.Decode(&tracksList)
	assertOk(err)
	for _,v := range tracksList.Data.Tracks {
		index := v.Index
		title := v.Title
		trackId := v.TrackId
		trackUrl := getTrackUrl(trackId)
		downTrack(index,trackUrl,title)
	}

	if tracksList.Data.TrackTotalCount > tracksList.Data.PageNum * tracksList.Data.PageSize {
		handleAlbumPage(albumId,pageNum+1)
	}
}

func getTrackUrl(trackId int) string{
	url := fmt.Sprintf("https://www.ximalaya.com/revision/play/v1/audio?id=%d&ptype=1",trackId)
	resp, err := http.Get(url)
	assertOk(err)
	//parse json
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var track Track
	err = decoder.Decode(&track)
	assertOk(err)
	return track.Data.Src
}

func downTrack(index int,trackUrl string,trackName string) {
	resp, err := http.Get(trackUrl)
	if err != nil {
		log.Info(err.Error())
		return
	}
	os.MkdirAll(ablumPath,os.ModePerm)
	itoa := strconv.Itoa(index)
	seq := lpad(itoa, 4)
	filename := seq+"_"+trackName + ".m4a"
	file, _ := os.Create(ablumPath+"/"+ filename)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("download %s success", filename)
	io.Copy(file,resp.Body)
}

func lpad(s string,n int) string{
	l := len(s)
	if(l >= n){
		return s
	}
	result := ""
	for i:=0;i < n-l ;i++  {
		result += "0"
	}
	result += s
	return result
}

type Album struct {
	Data struct{
		MainInfo struct{
			AlbumTitle string `json:"albumTitle"`
		} `json:"mainInfo"`

	} `json:"data"`
}

type TracksList struct {
	Ret int `json:"ret"`
	Data struct{
		TrackTotalCount int `json:"trackTotalCount"`
		Tracks[] struct{
			Index int `json:"index"`
			TrackId int `json:"trackId"`
			Title string `json:"title"`
		}  `json:"tracks"`
		PageNum  int `json:"pageNum"`
		PageSize int `json:"pageSize"`
	} `json:"data"`

}

type Track struct {
	Ret int `json:"ret"`
	Data struct{
		TrackId int `json:"trackId"`
		Src string `json:"Src"`
	} `json:"data"`
}

func assertOk(err error){
	if err != nil {
		panic(err)
	}
}

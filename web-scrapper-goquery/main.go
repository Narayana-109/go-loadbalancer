package main

import (
	"fmt"
	"log"
	"strconv"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)


func main(){
	blogTitles, err := GetLatestBlogTitles("https://go.dev")
	if err != nil{
		log.Println(err)
	}
	fmt.Println("Blog Titles:")
	fmt.Println(blogTitles)
}

func GetLatestBlogTitles(url string)(string, error){
	resp, err := http.Get(url)

	if err != nil{
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return "", err
	}
	titles := ""
	doc.Find(".WhyGo-reasonTitle").Each(func(i int, s *goquery.Selection){
		titles += strconv.Itoa(i+1) + "-" + s.Text() + "\n"
	})
	return titles, nil
}
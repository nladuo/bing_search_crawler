package main

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/levigross/grequests"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/url"
	"strings"
	"sync"
)

type Task struct {
	Id        bson.ObjectId `bson:"_id"`
	Qid       string        `bson:"qid"`
	Query     string        `bson:"query"`
	IsCrawled bool          `bson:"is_crawled"`
}

func parseBody(body string) []string {
	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(strings.NewReader(body)); err != nil {
		return []string{}
	}
	root := doc.SelectElement("rss")
	if root == nil {
		return []string{}
	}

	root = root.SelectElement("channel")
	if root == nil {
		return []string{}
	}
	var texts = []string{}
	for _, item := range root.SelectElements("item") {
		if desc := item.SelectElement("description"); desc != nil {
			// fmt.Printf("  desc: %s \n", desc.Text())
			texts = append(texts, desc.Text())
		}
	}
	return texts
}

func crawlAndUpdate(task Task, coll *mgo.Collection, counter int) error {
	v := url.Values{}
	v.Add("q", task.Query)
	query := v.Encode()
	url := "http://www.bing.com/search?format=rss&ensearch=1&FORM=QBLH&" + query
	failedTime := 0
TAG:
	res, err := grequests.Get(url, &grequests.RequestOptions{
		Headers: map[string]string{
			"Cookie":     "_SS=SID=1F6B2D15FEC8668E38B8206BFFB76702&HV=1560738706; _EDGE_V=1; SRCHUSR=DOB=20190617; _EDGE_S=F=1&SID=1F6B2D15FEC8668E38B8206BFFB76702; MUID=1742D66F483D663525F7DB1149426784; SRCHUID=V=2&GUID=07B9E056CB9E40B6A16F3FC3DE2DFA8B&dmnchg=1; MUIDB=1742D66F483D663525F7DB1149426784; SRCHD=AF=QBLH;",
			"User-Agent": "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36",
		},
	})
	if err != nil {
		return err
	}

	texts := parseBody(res.String())

	if (len(texts) == 0) && (failedTime <= 4) {
		failedTime += 1
		goto TAG
	}
	// fmt.Println(texts)

	fmt.Println(counter, task.Qid, task.Query, len(texts))

	data := bson.M{"$set": bson.M{"texts": texts, "is_crawled": true}}
	err = coll.Update(bson.M{"_id": task.Id}, data)
	return nil
}

var (
	taskChan chan Task
	addChan  chan int // channel for set the max goroutine count
)

func main() {
	var task Task
	taskChan = make(chan Task, 50)
	addChan = make(chan int, 15)
	var wg sync.WaitGroup

	session, err := mgo.Dial("127.0.0.1:27017")
	defer session.Close()

	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)
	c := session.DB("bing").C("tasks")

	tasks := c.Find(bson.M{"is_crawled": false}).Iter()

	go func() {
		for tasks.Next(&task) {
			taskChan <- task
		}
		taskChan <- Task{
			Qid: "close",
		}
	}()

	counter := 0
	for {
		counter += 1
		t := <-taskChan
		if t.Qid != "close" {
			wg.Add(1)
			addChan <- 1
			go func(i int) {
				defer wg.Done()
				crawlAndUpdate(t, c, i)
				<-addChan
			}(counter)
		} else {
			break
		}
	}
	fmt.Println("outside")
	wg.Wait()

}

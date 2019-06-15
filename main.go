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
	failedTime := 0
Tag:
	// fmt.Println(query)
	url := "http://www.bing.com/search?format=rss&ensearch=1&FORM=QBLH&" + query
	res, err := grequests.Get(url, &grequests.RequestOptions{
		Headers: map[string]string{
			"Cookie":     "SRCHD=AF=QBLH; SRCHUID=V=2&GUID=9E91686DABB746D3A145437AAA51664A&dmnchg=1;SRCHUSR=DOB=20190615;_EDGE_S=F=1&SID=0C4E80C2FD54606411D68DBEFC7A617B;_EDGE_V=1;_SS=SID=0C4E80C2FD54606411D68DBEFC7A617B&HV=1560621128;DUP=Q=nKXbmv1AbAz0eYEX9yQHqw2&T=361475528&A=2&IG=5C70812122AF4E0CB0E68E4C73522768;MUIDB=18D6EDD8BAD660C22019E0A4BBF861E7;",
			"User-Agent": "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36",
		},
	},
	)
	if err != nil {
		return err
	}

	texts := parseBody(res.String())

	if (len(texts) == 0) && (failedTime <= 4) {
		failedTime += 1
		goto Tag
	}
	// fmt.Println(texts)

	fmt.Println(totalCount, counter, task.Qid, task.Query, len(texts))
	totalCount++

	data := bson.M{"$set": bson.M{"texts": texts, "is_crawled": true}}
	err = coll.Update(bson.M{"_id": task.Id}, data)
	return nil
}

var (
	taskChan chan Task
	// updateChan chan []string
	addChan    chan int
	totalCount int
)

func main() {

	session, err := mgo.Dial("127.0.0.1:27017")
	defer session.Close()

	if err != nil {
		panic(err)
	}

	taskChan = make(chan Task, 50)
	addChan = make(chan int, 15) // the concurrent goroutine count

	session.SetMode(mgo.Monotonic, true)
	c := session.DB("bing").C("tasks")

	var task Task

	tasks := c.Find(bson.M{"is_crawled": false}).Iter()

	var wg sync.WaitGroup
	totalCount = 0

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

# bing_search_crawler
Bing search crawler for academic use.


## Dependency

``` shell
go get gopkg.in/mgo.v2
go get github.com/beevik/etree
go get github.com/levigross/grequests
```


## Usage
### create your queries
Insert the queries to search in your MongoDB, like:
``` bson
{
    "_id" : ObjectId("5d04e2ac9c6b976ae6f8bf4a"),
    "query" : "Which laptop is this used by Mark Zuckerberg?",
    "qid" : "96682",
    "is_crawled" : false
}
```

### get cookie with headless chrome
``` shell
python export_cookie.py
```

### run the crawler

Update the cookie param of request headers.
``` go
res, err := grequests.Get(url, &grequests.RequestOptions{
	Headers: map[string]string{
		"Cookie":     "_SS=SID=1F6B2D15FEC8668E38B8206BFFB76702&HV=1560738706; _EDGE_V=1; SRCHUSR=DOB=20190617; _EDGE_S=F=1&SID=1F6B2D15FEC8668E38B8206BFFB76702; MUID=1742D66F483D663525F7DB1149426784; SRCHUID=V=2&GUID=07B9E056CB9E40B6A16F3FC3DE2DFA8B&dmnchg=1; MUIDB=1742D66F483D663525F7DB1149426784; SRCHD=AF=QBLH;",
		"User-Agent": "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36",
	},
})
```

And run bing_crawler.go

``` shell
go run bing_crawler.go
```


## LICENSE
MIT
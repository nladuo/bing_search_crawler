# bing_search_crawler
Bing search crawler for academic use.


## Dependency
``` shell
go get gopkg.in/mgo.v2
go get github.com/beevik/etree
go get github.com/levigross/grequests
```


## Usage
Insert the queries to search in your MongoDB, like:
``` json
{
    "_id" : ObjectId("5d04e2ac9c6b976ae6f8bf4a"),
    "query" : "Which laptop is this used by Mark Zuckerberg?",
    "qid" : "96682",
    "is_crawled" : false
}
```
And run main.go

``` shell
go run main.go
```


## LICENSE
MIT
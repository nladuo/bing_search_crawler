from xml.etree import ElementTree
from urllib.parse import quote
from selenium import webdriver
import requests
import json
import pymongo
from multiprocessing import Pool

client = pymongo.MongoClient()
db = client.bing
tasks = db.tasks


def bing_search(session, query):
    while True:
        try:
            response = session.get("http://cn.bing.com/search?format=rss&ensearch=1&FORM=QBLH&q=%s" % quote(query), headers={
                "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
                "cookie": "_SS=SID=1F6B2D15FEC8668E38B8206BFFB76702&HV=1560738706; _EDGE_V=1; SRCHUSR=DOB=20190617; _EDGE_S=F=1&SID=1F6B2D15FEC8668E38B8206BFFB76702; MUID=1742D66F483D663525F7DB1149426784; SRCHUID=V=2&GUID=07B9E056CB9E40B6A16F3FC3DE2DFA8B&dmnchg=1; MUIDB=1742D66F483D663525F7DB1149426784; SRCHD=AF=QBLH;"
            }, timeout=3)
            # print(response.content)
            tree = ElementTree.fromstring(response.content)
            # tree = ElementTree.fromstring(driver.page_source)
            x = tree.find('item')
            texts = []
            for it in tree.find('channel').findall('item'):
                texts.append(it.find('description').text)
            break
        except:
            pass
    return texts


def get_new_session():
    print("get new session")
    options = webdriver.ChromeOptions()
    options.add_argument('--disable-extensions')
    options.add_argument('--headless')
    driver = webdriver.Chrome(chrome_options=options)

    session = requests.session()
    driver.get("https://cn.bing.com/search?q=hellp&qs=n&form=QBLH&sp=-1&pq=&sc=0-0&sk=&cvid=36D87CF616344F2EB887E777C3E0BB2F")
    for cookie in driver.get_cookies():
        session.cookies.set(cookie["name"], cookie["value"], domain=cookie["domain"], path=cookie["path"])

    driver.quit()
    return session

count = 0
session = get_new_session()


def crawl_one(task, count):
    texts = bing_search(session, task["query"])

    failed_time = 0
    while len(texts) == 0:
        print("zero length")
        texts = bing_search(session, task["query"])
        failed_time += 1
        if failed_time > 4:
            break

    print(count, task["qid"], task["query"], len(texts))
    tasks.update({'_id': task["_id"]}, {
        '$set': {
            "is_crawled": True,
            "texts": texts
        },
    })

pool = Pool(50)

for task in tasks.find({"is_crawled": False}):
    pool.apply_async(crawl_one, args=(task, count))
    count += 1

pool.close()
pool.join()

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
                "User-Agent": "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36",
                # "Cookie":     "DUP=Q=cyf2E319RplzT70uHnlIqw2&T=368615744&A=2&IG=FC18BC5497064931959EE9F6DA0B10D2; MUID=1AE2679CCD1F6E4533796B00C91F6D54; SRCHD=AF=NOFORM; MSCC=1; MUIDB=1AE2679CCD1F6E4533796B00C91F6D54; _EDGE_S=mkt=zh-cn&SID=24BE194AE40368593F351431E55B692C; _FP=hta=on; SerpPWA=reg=1; ULC=P=F0B3|1:1&H=F0B3|1:1&T=F0B3|1:1; SRCHHPGUSR=CW=1280&CH=689&DPR=2&UTC=480&WTS=63696304825; ipv6=hit=1560711627371&t=4; SRCHUSR=DOB=20190614&T=1560708028000; ENSEARCH=BENVER=0; _SS=SID=24BE194AE40368593F351431E55B692C&bIm=525816&HV=1560708033; SRCHUID=V=2&GUID=E36C49C8C20D43F6804EC01055B2921D&dmnchg=1; SNRHOP=I=&TS=",
            }, timeout=5)
            # print(response.content)
            body = response.content
            tree = ElementTree.fromstring(response.content)
            # tree = ElementTree.fromstring(driver.page_source)
            x = tree.find('item')
            texts = []
            for it in tree.find('channel').findall('item'):
                texts.append(it.find('description').text)
            break
        except:
            pass
    return texts, body


def get_new_session():
    print("get new session")
    options = webdriver.ChromeOptions()
    options.add_argument('--disable-extensions')
    options.add_argument('--headless')
    driver = webdriver.Chrome(chrome_options=options)

    session = requests.session()
    driver.get("https://cn.bing.com/search?q=hellp&qs=n&form=QBLH&sp=-1&pq=&sc=0-0&sk=&cvid=36D87CF616344F2EB887E777C3E0BB2F")
    print(driver.get_cookies())
    for cookie in driver.get_cookies():
        session.cookies.set(cookie["name"], cookie["value"], domain=cookie["domain"], path=cookie["path"])

    driver.quit()
    return session

count = 0
session = get_new_session()


def crawl_one(task, count):
    texts, body  = bing_search(session, task["query"])

    failed_time = 0
    while len(texts) == 0:
        print("zero length")
        texts, body = bing_search(session, task["query"])
        failed_time += 1
        if failed_time > 4:
            break
    print(task)
    print(count, task["query"], len(texts))
    tasks.update({'_id': task["_id"]}, {
        '$set': {
            "is_crawled": True,
            "body": body,
            "answer_count": len(texts)
        },
    })

pool = Pool(10)

for task in tasks.find({"is_crawled": False}):
    pool.apply_async(crawl_one, args=(task, count))
    # crawl_one(task, count)
    count += 1

pool.close()
pool.join()

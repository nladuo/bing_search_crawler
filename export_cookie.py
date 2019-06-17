from selenium import webdriver
import json

options = webdriver.ChromeOptions()
options.add_argument('--disable-extensions')
options.add_argument('--headless')
# options.add_argument('--disable-gpu')
# options.add_argument('--no-sandbox')
driver = webdriver.Chrome(chrome_options=options)

driver.get("https://cn.bing.com/search?q=hellp&qs=n&form=QBLH&sp=-1&pq=&sc=0-0&sk=&cvid=36D87CF616344F2EB887E777C3E0BB2F")

cookie_str = ""
for cookie in driver.get_cookies():
    cookie_str += cookie["name"] + "=" + cookie["value"] + "; "

driver.quit()

print(cookie_str)

import requests, json

appkey="1daf57127d394c638ba7f57c5287c162"
secret="9051a58e08d2a03dbdb4cd6f47e31c34"
token="at.6f7hp66j57x5vbeed6j1rgjo089jhcrd-3uu3jbhb3g-14fuq9t-eikm6scvd"

url="https://open.ys7.com/api/lapp/live/video/list"

data = {
    "POST": "api/lapp/token/get HTTP/1.1",
    "Host": "open.ys7.com",
    "Content-Type": "application/x-www-form-urlencoded",
    "accessToken": token
}

res = requests.post(url, data=data)
resJson = json.loads(res.text)
print(resJson)
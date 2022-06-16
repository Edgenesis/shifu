## 需要编写解码模块字典到configMap，例如:
```
......
data:
    decoding: |
        enabled : "true"
        decodeDictionaries:
            getContentJson:
                responseMsgType: "json"
            getcontentString:
                responseMsgType: "string"
            getcontentformatted: 
                responseMsgType: "formattedString"
                dictionaryDilimiter: ','
                dictionary:
                    - "age"
                    - "name"
                    - "gender"
......
```
``

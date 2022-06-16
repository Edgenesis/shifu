## The decoding module dictionary needs to be written to configmap, for example:
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


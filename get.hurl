GET http://localhost:8080/resource/1
HTTP 200

[Captures]
etag: header "ETag"

GET http://localhost:8080/resource/1
If-Match: {{etag}}
HTTP 304

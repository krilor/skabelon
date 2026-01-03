GET http://localhost:8080/resource/1
HTTP 200

[Captures]
correct_etag: header "ETag"

PATCH http://localhost:8080/resource/1
If-Match: "randometag"
{
    "rkey": "{{newUuid}}",
    "is_fun": false,
    "my_int" : 1,
    "description": "{{newDate}}"
}
HTTP 412

PATCH http://localhost:8080/resource/1
If-Match: {{correct_etag}}
{
    "rkey": "{{newUuid}}",
    "is_fun": false,
    "my_int" : 1,
    "description": "{{newDate}}"
}
HTTP 200

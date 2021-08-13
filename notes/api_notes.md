# Use 
- App code for our api
- running the server
- reading and writing http requests
- managing authentication
# Test the health of the app

```bash
$ curl -i localhost:4000/v1/healthcheck
HTTP/1.1 200 OK
Date: Fri, 13 Aug 2021 02:37:52 GMT
Content-Length: 52
Content-Type: text/plain; charset=utf-8

Status: available
environment:  dev
version:  1.0.0
```


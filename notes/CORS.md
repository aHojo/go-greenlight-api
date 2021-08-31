Basically, if two URLs have the same scheme, host and port (if specified) they are said to share the same origin. To help illustrate this, let’s compare the following URLs:

URL A      |       URL B Same       |      origin?       |        Reason
-----------|-----------|-----------|-----------|
https://foo.com/a  | http://foo.com/a     | No           | Different scheme (http vs https)
http://foo.com/a   | http://www.foo.com/a | No           | Different host   (foo.com vs www.foo.com)
http://foo.com/a   | http://foo.com:443/a | No           | Different port (no port vs 443)
http://foo.com/a   | http://foo.com/b     | Yes          | Only the path is different
http://foo.com/a   | http://foo.com/a?b=c | Yes          | Only the query string is different
http://foo.com/a#b | http://foo.com/a#c   | Yes          | Only the fragment is different

Understanding what origins are is important because all web browsers implement a security mechanism known as the same-origin policy. There are some very small differences in how browsers implement this policy, but broadly speaking:

- A webpage on one origin can embed certain types of resources from another origin in their HTML — including images, CSS, and JavaScript files. For example, doing this is in your webpage is OK:
`<img src="http://anotherorigin.com/example.png" alt="example image">`

- A webpage on one origin can send data to a different origin. For example, it’s OK for a HTML form in a webpage to submit data to a different origin.

- But a webpage on one origin is not allowed to receive data from a different origin. 

This key thing here is the final bullet-point: the same-origin policy prevents a (potentially malicious) website on another origin from reading (possibly confidential) information from your website

# Cors request is "simple" on these conditions
- The request HTTP method is one of the three CORS-safe methods: HEAD, GET or POST.
- The request headers are all either forbidden headers or one of the four CORS-safe headers:
  - Accept
  - Accept-Language
  - Content-Language
  - Content-Type
- The value for the Content-Type header (if set) is one of:
  - application/x-www-form-urlencoded
  - multipart/form-data
  - text/plain

When a cross-origin request doesn’t meet these conditions, then the web browser will trigger an initial ‘preflight’ request before the real request. 



There are three headers here which are relevant to CORS:
- `Origin` — As we saw previously, this lets our API know what origin the preflight request is coming from.
- `Access`-Control-Request-Method — This lets our API know what HTTP method will be used for the real request (in this case, we can see that the real request will be a POST).
- `Access-Control-Request-Headers` — This lets our API know what HTTP headers will be sent with the real request (in this case we can see that the real request will include a content-type header).


Once we identify that it is a preflight request, we need to send a 200 OK response with some special headers to let the browser know whether or not it’s OK for the real request to proceed. These are: 
- An `Access-Control-Allow-Origin` response header, which reflects the value of the preflight request’s Origin header (just like in the previous chapter).
- An `Access-Control-Allow-Methods` header listing the HTTP methods that can be used in real cross-origin requests to the URL.
- An `Access-Control-Allow-Headers` header listing the request headers that can be included in real cross-origin requests to the URL.

In our case, we could set the following response headers to allow cross-origin requests for all our endpoints:

```bash 
Access-Control-Allow-Origin: <reflected trusted origin>
Access-Control-Allow-Methods: OPTIONS, PUT, PATCH, DELETE
Access-Control-Allow-Headers: Authorization, Content-Type
```

Important: When responding to a preflight request it’s not necessary to include the
CORS-safe methods HEAD, GET or POST in the Access-Control-Allow-Methods header.
Likewise, it’s not necessary to include forbidden or CORS-safe headers in
Access-Control-Allow-Headers.



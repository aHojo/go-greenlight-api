# Basic Auth
The client includes an Authorization header with every request containing their credentials. The credentials need to be in the format `username:password` and `base-64` encoded. So, for example, to authenticate as `alice@example.com:pa55word` the client would send the following header:

```
Authorization: Basic YWxpY2VAZXhhbXBsZS5jb206cGE1NXdvcmQ=
```

You can then extract the credentials from this header using Go’s -
[Request.BasicAuth()](https://golang.org/pkg/net/http/#Request.BasicAuth) then verify they are correct

# Token Authentication

1. The client sends a request to your API containing their credentials (typically username or email address, and password).

2. The API verifies that the credentials are correct, generates a bearer token which represents the user, and sends it back to the user. The token expires after a set period of time, after which the user will need to resubmit their credentials again to get a new token.

3. For subsequent requests to the API, the client includes the token in an Authorization header like this:

Authorization: Bearer <token>

4. When your API receives this request, it checks that the token hasn’t expired and examines the token value to determine who the user is.

## Stateful Token Authentication
In a stateful token approach, the value of the token is a high-entropy cryptographicallysecure random string. This token — or a fast hash of it — is stored server-side in a database, alongside the user ID and an expiry time for the token.

When the client sends back the token in subsequent requests, your API can look up the token in the database, check that it hasn’t expired, and retrieve the corresponding user ID to find out who the request is coming from.

## Stateless Token Authentication

Stateless tokens encode the user ID and expiry time in the token itself

The token is cryptographically signed to prevent tampering and (in some cases) encrypted to prevent
the contents being read.

- JWT https://en.wikipedia.org/wiki/JSON_Web_Token
- PASETO https://developer.okta.com/blog/2019/10/17/a-thorough-introduction-to-paseto
- Branca https://branca.io/ 
- nacl/secretbox https://pkg.go.dev/golang.org/x/crypto/nacl/secretbox

# API-key Authentication

The idea behind API-key authentication is that a user has a non-expiring secret ‘key’ associated with their account. This key should be a high-entropy cryptographically-secure random string, and a fast hash of the key (SHA256 or SHA512) should be stored alongside the corresponding user ID in your database.

The user then passes their key with each request to your API in a header like this:  
```Authorization: Key <key>```

On receiving it, your API can regenerate the fast hash of the key and use it to lookup the corresponding user ID from your database.

# Authenticating Requests 
Once a client has an authentication token we will expect them to include it with all subsequent requests in an Authorization header, like so:

```Authorization: Bearer IEYZQUBEMPPAKPOAWTPV6YJ6RM```

- If the authentication token is not valid, then we will send the client a 401 Unauthorized response and an error message to let them know that their token is malformed or invalid.

-  If the authentication token is valid, we will look up the user details and add their details to the request context.

- If no Authorization header was provided at all, then we will add the details for an anonymous user to the request context instead.

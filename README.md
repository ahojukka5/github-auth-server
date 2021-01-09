# github-auth-server

Github authentication server to be used with react-github-auth.

To test with curl:

```bash
curl -X POST -H "Content-Type: application/json" \
     -D '{"code": "19da77673aa1f6e140da"}' \
     localhost:8080
```

Client id must be given as an environment variable `GITHUB_CLIENT_ID`. Client
secret must be given as an environment variable `GITHUB_CLIENT_SECRET`.

Response is of type application/json, having `name`, `email` and `token`.

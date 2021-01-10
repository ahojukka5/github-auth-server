# github-auth-server

github-auth-server is a small, lightweight Go server to authenticate using
GitHub OAuth authentication service. Server takes json payload with `code`
defined in it, sends it to GitHub authentication, and if authentication is
succesful, it returns some minimal user information that can be used to identify
user, and additionally GitHub access token. `GITHUB_CLIENT_ID` and
`GITHUB_CLIENT_SECRET` must be defined. For more information on how to get them,
take a look of another package `react-github-auth` and it's readme[1][1].

[1]: https://github.com/ahojukka5/react-github-auth

To test with curl:

```bash
curl localhost:8080/authenticate/github?code=19da77673aa1f6e140da
```

Client id must be given as an environment variable `GITHUB_CLIENT_ID`. Client
secret must be given as an environment variable `GITHUB_CLIENT_SECRET`.

Response is of type application/json, having `name`, `email` and `token`.

## Using with Docker

```bash
docker run -d --name=github-auth-server \
        --env GITHUB_CLIENT_ID=get_this_from_github \
        --env GITHUB_CLIENT_SECRET=get_this_from_github \
        -p 8080:8080 ahojukka5/github-auth-server
```

After that, run `curl` like above.

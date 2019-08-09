# OAuth Bridge

This service redirects the user to a given frontend application with a valid access
token as a parameter in the url.

## Setup

1. Create a `config.json` file to define settings and OAuth services to host on the server.

```
{
  "redirect_uri": "[url and port of web server]",
  "port": "[port of web server]",
  "services": {
    ...
  }
}
```

Each service is an API that requires OAuth for authentication. You can define the application specific settings for each service:

```
"services": {
  "spotify": {
    "client_id": "[OAuth application client id]",
    "client_secret": "[OAuth application client secret]",
    "redirect_uri": "https://accounts.spotify.com/authorize",
    "scope": "[required oauth scopes]"
  }
}
```

This allows the application to be used as a generic OAuth authenication server.
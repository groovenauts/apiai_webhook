# magellan-gcs-uploader

## Run local server

```
goapp serve
```

## Deploy

Specify GCP project id and api tokens (comma separated).

```
appcfg.py -E API_TOKEN:XXXXXXXX -E BLOCKS_URL:https://xxxx.magellanic-clouds.net/ -E BLOCKS_API_TOKEN:xxxxxx update .
```

### Environment Variables

| Name | Required | Description |
|------|----------|-------------|
| `API_TOKEN` | o | API Tokens (comma separated) to be compared with request's key param for authorization |
| `BLOCKS_URL` | x | Hook URL to invoke BLOCKS flow. |
| `BLOCKS_API_TOKEN` | x | API Tokent for BLOCKS Board. |

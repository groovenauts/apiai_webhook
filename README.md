# apiai_webhook

## Run local server

```
goapp serve
```

## Deploy

Specify GCP project id and api tokens (comma separated).

```
cp app.yaml.example app.yaml
(Edit app.yaml to setup environments)
gcloud --project ${PROJECT_ID} app deploy app.yaml -v v1
```

### Environment Variables

| Name | Required | Description |
|------|----------|-------------|
| `API_TOKEN` | o | API Tokens (comma separated) to be compared with request's key param for authorization |
| `BLOCKS_URL` | x | Hook URL to invoke BLOCKS flow. |
| `BLOCKS_API_TOKEN` | x | API Tokent for BLOCKS Board. |

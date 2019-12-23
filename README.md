# deploy-webhook

Deploy from Github/Travis CI to Server using Webhook API

## Setup for testing

To test the service from Github or other provider we need to have a publicly accessible server.
If you already have a public webhook endpoint you can skip this step.

### Install localtunnel

```bash
sudo npm install -g localtunnel
```

Start tunnel after the application have been started

```bash
lt --port 8080 --subdomain deploy-hook
```

Accessing service via deployed tunnel

```bash
https://deploy-hook.localtunnel.me/deploy
```

## Compiling and Running

Application is written in `golang` so you will need that installed first.

```bash
# compile and run
go run . 127.0.0.7:8787

# build executable
go build

# build executable and launch application
go build && ./deploy-webhook 127.0.0.1:8787
```

Starting application

```bash
# Usage
./deploy-webhook host:port

# Bind to loopback interface on port 8787
./deploy-webhook 127.0.0.1:8787

# Bind to any interface on port 8787
./deploy-webhook :8787
```

There will be couple endpoint that are eposed after the application is started

```bash
GET  /         : Display status of the service
POST /deploy   : Handle incomming payload
GET  /metrics  : Display deployment metrics
```

Status request `127.0.0.1:8787`

Response

```bash
Webhook Service : Dec 21 22:44:55
```

Metrics request `127.0.0.1:8787/metrics`

Response

```json
{ 
   "project-name":{ 
      "hits":1,
      "commit":"a09c01b8cefff3d7cb831c13c3551d9bc358a7f1",
      "ref":"refs/heads/master",
      "time":"Dec 21 23:43:41",
      "duration":250
   }
}
```

## systemd setup

Copy the sevice file `deploy-webhook.service` to `/etc/systemd/system`

```bash
sudo cp deploy-webhook.service /etc/systemd/system/deploy-webhook.service
```

Starting the Service

```bash
sudo systemctl start deploy-webhook.service
```

Check the Service Status

```bash
sudo systemctl status deploy-webhook.service
```

View the logs

```bash
journalctl -u deploy-webhook -e
```

## Tools

JSON to GO struct conversion

https://developer.github.com/v3/activity/events/types/#pushevent
https://mholt.github.io/json-to-go/

During conversion there are couple issues where we have to change some fields from `int` to `time.Time` in generated code.

## Notes

Added support handling both content-types `application/json` and `application/x-www-form-urlencoded`
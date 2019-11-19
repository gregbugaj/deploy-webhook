# deploy-webhook
Deploy from Github/Travis CI to Server using Webhook API

# Setup for testing 

Install localtunnel

```
sudo npm install -g localtunnel
```

Start tunnel after the application have been started 
```
lt --port 8080 --subdomain deploy-hook
```

Access service via deployed tunnel

```
https://deploy-hook.localtunnel.me/deploy-hook
```

# JSON to GO struct conversion
https://developer.github.com/v3/activity/events/types/#pushevent
https://mholt.github.io/json-to-go/

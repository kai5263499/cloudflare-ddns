# cloudflare-ddns
This is a simple utility to create or update an A record on a CloudFlare-managed domain and have it point to the external IP of wherever the container is run.

In other words, it allows you to setup a custom dynamic DNS provider.

The real value in this utility is when it's run as a docker service with a regular update interval.

# Usage

```
docker run --rm \
-e CF_API_KEY=secret \
-e CF_API_EMAIL=account@cloudflare.com \
-e ZONE=top-domain.com \
-e NAME=subdomain \
-e UPDATE_INTERVAL=10 \
--name ddns_subdomain \
kai5263499/cloudflare-ddns
```
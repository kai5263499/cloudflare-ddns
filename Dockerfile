FROM golang:alpine as gobuilder

RUN apk --no-cache add ca-certificates curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    mkdir -p /go/src/github.com/kai5263499/cloudflare-ddns
COPY . /go/src/github.com/kai5263499/cloudflare-ddns
WORKDIR /go/src/github.com/kai5263499/cloudflare-ddns/cmd/cloudflare-ddns
RUN dep ensure && CGO_ENABLED=0 go build -a -o cloudflare-ddns

FROM scratch

COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuilder /go/src/github.com/kai5263499/cloudflare-ddns/cmd/cloudflare-ddns/cloudflare-ddns /cloudflare-ddns

ENV CF_API_KEY=""
ENV CF_API_EMAIL=""
ENV ZONE=""
ENV NAME=""
ENV UPDATE_INTERVAL=300

ENTRYPOINT [ "/cloudflare-ddns" ]
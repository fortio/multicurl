FROM alpine as certs
RUN apk update && apk add ca-certificates
FROM scratch
COPY multicurl /usr/bin/multicurl
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/usr/bin/multicurl"]

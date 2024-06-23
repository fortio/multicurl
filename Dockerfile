# We don't need to copy the CA bundle anymore thanks to
# https://github.com/fortio/cli/releases/tag/v1.6.0
FROM scratch
COPY multicurl /usr/bin/multicurl
ENTRYPOINT ["/usr/bin/multicurl"]

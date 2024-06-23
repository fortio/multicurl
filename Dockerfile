FROM scratch
COPY multicurl /usr/bin/multicurl
ENTRYPOINT ["/usr/bin/multicurl"]

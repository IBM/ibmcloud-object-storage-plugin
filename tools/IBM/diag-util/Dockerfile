FROM alpine:3.7
RUN apk add --update bash
RUN apk add --update python
COPY entrypoint .
COPY check-mount-health .
RUN chmod u+x check-mount-health entrypoint

CMD ["./entrypoint"]

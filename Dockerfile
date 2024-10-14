FROM alpine:latest
WORKDIR /app
ADD bin/admission-webhook-server /app/admission-webhook-server
ENTRYPOINT ["./admission-webhook-server"]
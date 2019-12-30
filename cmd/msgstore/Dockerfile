# Stage 1

FROM golang:alpine as builder
RUN apk add --update go git build-base
RUN mkdir -p /build
ENV BUILDPATH /go/src/github.com/ghsbhatia/msgbox
RUN mkdir -p $BUILDPATH
WORKDIR $BUILDPATH
ADD . $BUILDPATH
RUN go build -o /build/msgstoresvc cmd/msgstore/msgstoreservice.go

# Stage 2

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/msgstoresvc /app/
WORKDIR /app
ENV MONGODB_URL mongodb://root:secret@127.0.0.1:27017/msgbox-mongo?authSource=admin&gssapiServiceName=mongodb
ENV USERSVC_URL http://localhost:6060
EXPOSE 6080
CMD ["./msgstoresvc"]

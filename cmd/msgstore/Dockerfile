# Stage 1

FROM golang:alpine as builder
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
EXPOSE 6080
CMD ["./msgstoresvc"]
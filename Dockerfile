FROM golang:latest AS BUILDER

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
ADD . .
RUN CGO_ENABLED=0 GO111MODULE=on go build -o clipper-api

FROM alpine:3.8

COPY --from=builder /go/src/app/clipper-api /bin/clipper-api

EXPOSE 8080
ENTRYPOINT ["/bin/clipper-api"]
CMD ["start"]

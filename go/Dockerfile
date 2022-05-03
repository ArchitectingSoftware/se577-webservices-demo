FROM golang:latest as builder

WORKDIR /go/src/drexel.edu/bc-service/go
COPY    go* ./
COPY    src ./src
RUN     CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v -o bin/bcservice ./src

FROM alpine
WORKDIR /bin/
COPY --from=builder /go/src/drexel.edu/bc-service/go/bin ./
ENTRYPOINT ["/bin/bcservice"]
EXPOSE  9095

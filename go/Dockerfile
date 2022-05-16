FROM golang:latest as builder
LABEL stage=builder
#PHASE 1:  Build the GO Binary - aka builder container
WORKDIR /go/src/drexel.edu/bc-service/go
COPY    go* ./
COPY    src ./src
RUN     CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v -o bin/bcservice ./src

#PHASE2: Build the final container, notice the directory naming follows the package prefix in go.mod
FROM alpine
WORKDIR /bin/
COPY --from=builder /go/src/drexel.edu/bc-service/go/bin ./
ENTRYPOINT ["/bin/bcservice"]
EXPOSE  9095

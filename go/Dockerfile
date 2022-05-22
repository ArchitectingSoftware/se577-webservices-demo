FROM golang:alpine as builder
LABEL stage=builder
#PHASE 1:  Build the GO Binary - aka builder container
WORKDIR /go/src/drexel.edu/bc-service/go
COPY    go* ./
COPY    src ./src
RUN     go mod download
RUN     go build -v -o bin/bcservice ./src
#RUN     CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v -o bin/bcservice ./src

#PHASE2: Build the final container, notice the directory naming follows the package prefix in go.mod
FROM alpine
WORKDIR /app
COPY --from=builder /go/src/drexel.edu/bc-service/go/bin ./
#the following create a non root user to run the app
RUN addgroup -S appgroup && adduser -S goacct -G appgroup
RUN chown -R goacct:root /app
USER goacct
ENTRYPOINT ["/app/bcservice"]
EXPOSE  9095

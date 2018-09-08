FROM golang:1.11.0-alpine as build
ENV GO111MODULE=on

RUN apk add --update --no-cache build-base git

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go .
RUN go build -o /bin/node-connector .

FROM alpine:latest

COPY --from=build /bin/node-connector /bin/node-connector 

ENTRYPOINT ["/bin/node-connector"]
EXPOSE 8080

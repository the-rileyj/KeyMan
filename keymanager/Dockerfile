FROM golang:1.11.3-alpine3.8

EXPOSE 9902

RUN apk add git

WORKDIR /go/src/github.com/the-rileyj/KeyMan/keymanager

RUN mkdir ./keymanaging

COPY ./main.go .
COPY ./keymanaging/keymanaging.go ./keymanaging

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["keymanager"]
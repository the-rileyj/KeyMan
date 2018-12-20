FROM golang:1.11.3-alpine3.8

RUN apk add git

EXPOSE 9901

WORKDIR /go/src/github.com/the-rileyj/KeyMan/gatekeeper

RUN mkdir ./gatekeeping

COPY main.go .
COPY ./gatekeeping/gatekeeping.go ./gatekeeping

RUN go get -d -v ./...
RUN go install -v ./...

COPY RJsecret.key .
COPY RJcert.crt .

CMD ["gatekeeper"]
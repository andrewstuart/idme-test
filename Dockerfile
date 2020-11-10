FROM golang:stretch

WORKDIR $GOPATH/src/app

ENTRYPOINT /app

ADD app /

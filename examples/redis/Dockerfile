FROM iron/go:dev

ARG FUNCPKG

ADD . $GOPATH/src/$FUNCPKG
WORKDIR $GOPATH/src/$FUNCPKG

RUN go get .
RUN go build -o func

ENTRYPOINT ["./func"]
FROM golang:1.21.1

RUN mkdir /juhena-forum

ADD . /juhena-forum

WORKDIR /juhena-forum
RUN go mod tidy
RUN go build -o my-forum-app data.go

CMD ["/juhena-forum/my-forum-app"]

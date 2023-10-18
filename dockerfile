FROM golang:1.21.1

RUN mkdir /forum

ADD . /forum

WORKDIR /forum
RUN go mod tidy
RUN go build -o my-forum-app data.go

CMD ["/forum/my-forum-app"]

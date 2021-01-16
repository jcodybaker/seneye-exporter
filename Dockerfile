FROM golang:1.15-alpine

WORKDIR /go/src/github.com/jcodybaker/seneye-exporter

COPY . .

RUN go get -d -v ./...
RUN go install -v ./cmd/seneye-exporter

CMD ["seneye-exporter"]

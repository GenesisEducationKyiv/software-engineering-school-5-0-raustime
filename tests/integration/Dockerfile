FROM golang:1.23

WORKDIR /tests

COPY ./tests/integration/go.mod ./
COPY ./tests/integration/go.sum ./
RUN go mod download

COPY . .

CMD go test -v ./tests/integration/...

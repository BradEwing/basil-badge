FROM golang:1.24-alpine

WORKDIR /root/

COPY go.mod go.sum api.go main.go ./
RUN go mod download

COPY main.go ./
RUN go build -o basil-badge .

EXPOSE 3000

CMD ["./basil-badge"]
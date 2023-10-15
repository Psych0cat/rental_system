FROM golang:alpine
WORKDIR /app
ADD . /app
WORKDIR /app/cmd/car-rental
RUN go build -o ../../bin/car-rental ./
EXPOSE 8080
WORKDIR /app
CMD ["./bin/car-rental"]
FROM golang:1.20

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change

COPY ./server.go ./
COPY ./auth/ ./auth
COPY ./register/ ./register
COPY ./go.mod ./
COPY ./go.sum ./
RUN ls

RUN go mod download && go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY . .
COPY --from=0 /usr/src/app/app ./
CMD ["./app"]


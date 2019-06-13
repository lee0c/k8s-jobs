from golang as build

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch as binary
COPY --from=builder ./mycert.crt /usr/local/share/ca-certificates/mycert.crt
COPY --from=builder manager /manager

ENTRYPOINT["/manager"]
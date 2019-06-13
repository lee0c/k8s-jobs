from golang as build

WORKDIR /manager

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch
COPY --from=builder /manager /manager

ENTRYPOINT["/manager"]
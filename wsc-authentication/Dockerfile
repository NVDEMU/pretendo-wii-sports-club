FROM golang:1.27.2-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server . \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/account-admin ./cmd/account-admin

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server ./server
COPY --from=build /out/account-admin ./account-admin
EXPOSE 60014/udp 8080/tcp
ENTRYPOINT ["./server"]

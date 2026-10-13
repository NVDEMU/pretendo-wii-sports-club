FROM golang:1.27.2-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server . \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server ./server
# Ensure a newly-created named volume inherits write permissions for nonroot.
COPY --chown=nonroot:nonroot --from=build /out/data /data
EXPOSE 60015/udp
ENTRYPOINT ["./server"]

FROM golang:alpine AS builder
WORKDIR /build

COPY . .
RUN go build -ldflags "-s -w" -o Dockport

FROM scratch
WORKDIR /Dockport

COPY --from=builder /build/Dockport .
ENTRYPOINT ["./Dockport"]

FROM golang:1.25-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/goimportsruler ./

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/goimportsruler /usr/local/bin/goimportsruler
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

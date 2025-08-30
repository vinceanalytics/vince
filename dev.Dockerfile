FROM golang:1.24 AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v

RUN CGO_ENABLED=0 go build -o /go/bin/vince

FROM gcr.io/distroless/static-debian12
LABEL org.opencontainers.image.authors="Geofrey Ernest"
LABEL org.opencontainers.image.source="https://github.com/vinceanalytics/vince"
LABEL org.opencontainers.image.documentation="https://vinceanalytics.com"
LABEL org.opencontainers.image.vendor="Geofrey Ernest"
LABEL org.opencontainers.image.description="The Cloud Native Web Analytics Platform."
LABEL org.opencontainers.image.licenses="AGPL-3.0"

COPY --from=build /go/bin/vince /
ENTRYPOINT ["/vince"]

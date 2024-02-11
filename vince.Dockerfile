FROM scratch
LABEL org.opencontainers.image.authors="Geofrey Ernest"
LABEL org.opencontainers.image.source="https://github.com/vinceanalytics/vince"
LABEL org.opencontainers.image.documentation="https://vinceanalytics.com"
LABEL org.opencontainers.image.vendor="Geofrey Ernest"
LABEL org.opencontainers.image.description="API first high performance self hosted and cost effective privacy friendly web analytics  server for organizations of any size"
LABEL org.opencontainers.image.licenses="Apache-2.0"
ENTRYPOINT ["/vince"]
COPY vince /
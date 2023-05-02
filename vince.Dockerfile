FROM scratch
LABEL org.opencontainers.image.authors="Geofrey Ernest"
LABEL org.opencontainers.image.url="https://vinceanalytics.com"
LABEL org.opencontainers.image.documentation="https://vinceanalytis.com/docs"
LABEL org.opencontainers.image.vendor="Geofrey Ernest"
LABEL org.opencontainers.image.description="The Cloud Native Web Analytics Platform."
LABEL org.opencontainers.image.licenses="AGPL-3.0"
ENTRYPOINT ["/vince"]
COPY vince /
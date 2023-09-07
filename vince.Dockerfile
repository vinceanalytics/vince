FROM scratch
LABEL org.opencontainers.image.authors="Geofrey Ernest"
LABEL org.opencontainers.image.source="https://github.com/vinceanalytics/vince"
LABEL org.opencontainers.image.documentation="https://vinceanalytics.github.io/guide/"
LABEL org.opencontainers.image.vendor="Geofrey Ernest"
LABEL org.opencontainers.image.description="The Cloud Native Web Analytics Platform."
LABEL org.opencontainers.image.licenses="Apache-2.0"
ENTRYPOINT ["/vince"]
COPY vince /
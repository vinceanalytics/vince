FROM scratch
LABEL org.opencontainers.image.authors="Geofrey Ernest"
LABEL org.opencontainers.image.url="https://vinceanalyti.co"
LABEL org.opencontainers.image.documentation="https://vinceanalyti.co/docs"
LABEL org.opencontainers.image.vendor="The Vince Analytics Team"
LABEL org.opencontainers.image.description="Self hosted web analytics platform"
LABEL org.opencontainers.image.licenses="AGPL-3.0"
ENTRYPOINT ["/vince"]
COPY vince /
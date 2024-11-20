getent passwd otelcol-otlp >/dev/null || useradd --system --user-group --no-create-home --shell /sbin/nologin vince

mkdir -p /var/lib/vince-data
chown -R vince:vince /var/lib/vince-data

export SITE_ID=vinceanalytics.com
export TOKEN=xxxx

curl "http://localhost:8080/api/v1/stats/timeseries?site_id=$SITE_ID&period=6mo" \
  -H "Authorization: Bearer ${TOKEN}"
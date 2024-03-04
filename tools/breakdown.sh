export SITE_ID=vinceanalytics.com
export TOKEN=xxxx

curl "http://localhost:8080/api/v1/stats/breakdown?site_id=$SITE_ID&period=6mo&property=source&metrics=visitors,bounce_rate&limit=5" \
  -H "Authorization: Bearer ${TOKEN}"

export SITE_ID=vinceanalytics.com
export TOKEN=xxxx

curl "http://localhost:8080/api/v1/stats/aggregate?site_id=$SITE_ID&period=6mo&metrics=visitors,pageviews,bounce_rate,visit_duration" \
  -H "Authorization: Bearer ${TOKEN}"
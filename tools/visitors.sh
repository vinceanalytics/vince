export SITE_ID=vinceanalytics.com
export TOKEN=xxxx

curl  "http://localhost:8080/api/v1/stats/realtime/visitors?site_id=$SITE_ID" \
  -H "Authorization: Bearer ${TOKEN}"
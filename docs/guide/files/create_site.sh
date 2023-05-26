curl -X 'POST' \
	-d '{"domain":"vince.example.com"}' \
	-H 'Authorization: Bearer $VINCE_BOOTSTRAP_KEY' 'http://localhost:8080/api/v1/sites'
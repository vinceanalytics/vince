dev:
	go build -o bin/vince
	VINCE_ADMIN_NAME=acme \
	VINCE_ADMIN_EMAIL=trial@vinceanalytics.com \
	VINCE_ADMIN_PASSWORD=1234 \
	VINCE_DOMAINS=vinceanalytics.com \
	VINCE_LICENSE=trial_license_key  ./bin/vince serve


css:
	cd assets && npm run css && cd -

major:
	go run ./internal/version/bump/main.go major
minor:
	go run ./internal/version/bump/main.go minor 
patch:
	go run ./internal/version/bump/main.go patch

view : bin/views
	./bin/views

bin/views: tools/views/main.go
	go build -o bin/views ./tools/views
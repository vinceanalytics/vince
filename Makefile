dev:
	go build -o bin/vince
	./bin/vince serve  --adminName yolo --adminPassword 1234 --domains vinceanalytics.com --profile

docker:
	docker run --rm  -p 8080:8080 -v ./vince-data:/vince-data ghcr.io/vinceanalytics/vince:v1.7.1 serve  --adminName acme --adminPassword 1234 --domains vinceanalytics.com --profile


css:
	cd assets && npm run css && cd -

scripts:
	cd tracker && npm run deploy && cd -

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
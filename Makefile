dev:
	go build -o bin/vince
	./bin/vince  --adminName=acme --adminEmail=acme@user.test \
	--adminPassword=1234

css:
	cd assets && npm run css && cd -

major:
	go run ./internal/version/bump/main.go major
minor:
	go run ./internal/version/bump/main.go minor 
patch:
	go run ./internal/version/bump/main.go patch
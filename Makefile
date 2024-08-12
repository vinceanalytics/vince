dev:
	go run main.go -admin.bootstrap -admin.name=acme -admin.email=acme@user.test \
	-admin.password=1234

css:
	cd assets && npm run css && cd -
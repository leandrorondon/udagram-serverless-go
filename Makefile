.PHONY: build clean deploy gomodgen

build: gomodgen
	export GO111MODULE=on
	# Feed
	env GOOS=linux go build -ldflags="-s -w" -o bin/feed/getFeed lambda/feed/getFeed/getFeed.go
	# User
	env GOOS=linux go build -ldflags="-s -w" -o bin/users/createUser lambda/users/createUser/createUser.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/users/login lambda/users/login/login.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh

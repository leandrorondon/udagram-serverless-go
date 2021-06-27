.PHONY: build clean deploy gomodgen

build: gomodgen
	export GO111MODULE=on
	# User
	env GOOS=linux go build -ldflags="-s -w" -o bin/feed/getFeed lambda/feed/getFeed/getFeed.go
	# Feed
	env GOOS=linux go build -ldflags="-s -w" -o bin/users/createUser lambda/users/createUser/createUser.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh

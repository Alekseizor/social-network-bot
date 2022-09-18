PWD = ${CURDIR}
NAME = social-network-bot

.PHONY: run
run:
	go run $(PWD)/cmd/$(NAME)/


.PHONY: build
build:
	go build -o bin/$(NAME) $(PWD)/cmd/$(NAME)

.PHONY: test
test:
	go test $(PWD)/... -parallel=10 -coverprofile=cover.out

.PHONY: local
local:
	cp .dist.env .env && cp config/config.stg.toml config/config.toml

# todo: Мок миграций для теста в последствии выпилим 
.PHONY: migrate
migrate:
	echo "No migrations implemented"


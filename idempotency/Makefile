install-tools:
	brew install watchman
	go install  github.com/rakyll/gotest@latest

watch:
	@watchman-make -p '*' -t tests

tests:
	clear
	@gotest -v

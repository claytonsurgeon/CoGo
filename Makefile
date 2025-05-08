
live:
	make live/server

live/server:
	go run github.com/air-verse/air@latest \
	--build.cmd "go build -o ./tmp/main ./main.go" \
	--build.bin "./tmp/main" \
	--build.args_bin "-addr=:4321" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go, templ" \
	--misc.clean_on_exit true

main:
	go build -ldflags '-w -s' -trimpath

debug:
	go build -gcflags=all="-N -l"

run:
	go run .

clean:
	go clean

main:
	go build -trimpath

debug:
	go build -gcflags=all="-N -l"

run:
	go run .

clean:
	go clean

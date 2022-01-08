build:
	env GOOS=linux GOARCH=amd64 go build
	gzip mixin-snapshots
	scp mixin-snapshots.gz snapshot:/home/one/snapshots
	rm -rf mixin-snapshots.gz
	ssh snapshot "cd /home/one/snapshots && rm -rf mixin-snapshots && gzip -d mixin-snapshots.gz && sudo systemctl restart snapshot-draw.service"
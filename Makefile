.PHONY: gomodgen deploy delete

gomodgen:
	GO111MODULE=on go mod init

deploy:
	gcloud functions deploy talipapops --entry-point Hello --runtime go113 --trigger-http

delete:
	gcloud functions delete talipapops --entry-point Hello --runtime go113 --trigger-http

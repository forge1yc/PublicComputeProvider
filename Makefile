# linux文件
#build_mac:
	#go build -o bin/mac_bounty_email cmd/main_email.go

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/go_linux_provider cmd/main.go

build_mac:
	go build -o bin/go_mac_provider cmd/main.go

build_win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/go_win_provider cmd/main.go

#build_win:
#	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/win_bounty_eamil cmd/main_email.go

buildEmailAll:
	go build -o bin/go_mac_provider cmd/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/go_linux_provider cmd/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/go_win_provider cmd/main.go

#buildCentosAll:
#	go build -o bin/mac_bounty_centos cmd/main_centos.go


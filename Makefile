.PHONY: linux docker

docker:
	docker run -p 6306:3306 --name crawler-mysql -e MYSQL_ROOT_PASSWORD=root -d mysql:5 

linux:
	GOOS=linux GOARCH=amd64 go build -o build/lenkrr .

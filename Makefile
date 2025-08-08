lint:
		golangci-lint run > golangci.yml		
build:
		docker build -t app . 
dockerup:
		docker-compose up
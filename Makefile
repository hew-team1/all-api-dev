api_build:
	docker build -t userapi:$(version) ./api/

api_start:
	docker run --tty -it -d -p 8080:8080 userapi:${version}

build:
	docker compose build

start:
	docker compose up -d

restart:
	docker compose down; \
    docker compose build; \
    docker compose up -d

dlog:
	docker logs $(co)_api

end_user_create:
	docker compose run awscli \
    --endpoint-url http://dynamodb:8000 \
    dynamodb create-table \
        --table-name EndUsers \
        --attribute-definitions \
            AttributeName=uid,AttributeType=S \
        --key-schema AttributeName=uid,KeyType=HASH \
        --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

recruit_create:
	docker compose run awscli \
    --endpoint-url http://dynamodb:8000 \
    dynamodb create-table \
        --table-name Recruits \
        --attribute-definitions \
            AttributeName=id,AttributeType=N \
        --key-schema AttributeName=id,KeyType=HASH \
        --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1




incr_create:
	docker-compose run awscli \
	--endpoint-url http://dynamodb:8000 \
	dynamodb create-table \
		--table-name AtomicCounter \
		--attribute-definitions \
			AttributeName=countKey,AttributeType=S \
		--key-schema AttributeName=countKey,KeyType=HASH \
		--provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
		&& \
	docker-compose run awscli \
	--endpoint-url http://dynamodb:8000 \
	dynamodb put-item \
		--table-name AtomicCounter  \
		--item \
			'{"countKey": {"S": "Recruits"}, "countNumber": {"N": "0"}}' \
		--return-consumed-capacity TOTAL
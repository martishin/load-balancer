build:
	docker build -t load-balancer .

run:
	docker-compose up -d

stop:
	docker-compose stop

stop-web:
	docker-compose stop web1 web2

remove:
	docker-compose down

logs:
	docker-compose logs -f

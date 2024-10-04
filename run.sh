#docker stop $(docker ps -q)
docker-compose down -v
docker-compose up --build

#!/bin/bash
reso_addr='637423626125.dkr.ecr.us-east-2.amazonaws.com/simplechatter/user-api-dev'
tag='latest'

container_name="simplechatter-user-api-test"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}

docker run -p 8888:8888  --name=${container_name} -d ${reso_addr}:${tag}

docker logs ${container_name}


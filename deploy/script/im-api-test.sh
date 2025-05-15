#!/bin/bash
reso_addr='637423626125.dkr.ecr.us-east-2.amazonaws.com/simplechatter/im-api-dev'
tag='latest'

container_name="simplechatter-im-api-test"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}

docker run -p 8882:8882  --name=${container_name} -d ${reso_addr}:${tag}
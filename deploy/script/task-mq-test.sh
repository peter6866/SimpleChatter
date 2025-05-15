#!/bin/bash
reso_addr='637423626125.dkr.ecr.us-east-2.amazonaws.com/simplechatter/task-mq-dev'
tag='latest'

container_name="simplechatter-task-mq-test"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}

docker run --name=${container_name} -d ${reso_addr}:${tag}

#!/bin/bash
reso_addr='637423626125.dkr.ecr.us-east-2.amazonaws.com/simplechatter/im-ws-dev'
tag='latest'

container_name="simplechatter-im-ws-test"

pod_ip="192.168.1.106"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}

docker run -p 10090:10090 -e POD_IP=${pod_ip} --name=${container_name} -d ${reso_addr}:${tag}

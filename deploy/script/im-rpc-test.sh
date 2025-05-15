#!/bin/bash
reso_addr='637423626125.dkr.ecr.us-east-2.amazonaws.com/simplechatter/im-rpc-dev'
tag='latest'

container_name="simplechatter-im-rpc-test"

pod_ip="192.168.1.106"

docker stop ${container_name}

docker rm ${container_name}

docker rmi ${reso_addr}:${tag}

docker pull ${reso_addr}:${tag}

docker run -p 10002:10002 -e POD_IP=${pod_ip} --name=${container_name} -d ${reso_addr}:${tag}

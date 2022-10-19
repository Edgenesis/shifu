#!bin/bash

default='empty'

for i in {1..5} 
do
    docker run -it --network host edgehub/mockclient:nightly

    output=$(docker exec -it nginx curl localhost:17773/data)
    echo $output
    if [ $output != $default ];then
        exit 0
    fi
done
exit 1

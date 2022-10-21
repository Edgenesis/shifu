#!bin/bash

default='empty'

for i in {1..100} 
do
    docker run -itd --network host edgehub/mockclient:nightly

    output=$(docker exec nginx curl localhost:17773/data)
    echo $output
    echo $default
    if [[ $output != $default ]]
    then
        exit 0
    fi
done
exit 1

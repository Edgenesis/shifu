#!bin/bash

default='empty'

for i in {1..5} 
do
    docker run --network host edgehub/mockclient:nightly

    output=$(docker exec -it nginx curl localhost:17773/data)
    echo $output
    if [[ $output -ne $default ]]
    then
        exit 0
    fi
done
exit 1

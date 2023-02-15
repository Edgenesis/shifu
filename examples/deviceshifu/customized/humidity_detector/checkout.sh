#!bin/bash

cleaned_raw_data="$(cat cleaned_raw_data)"
raw_data="$(cat raw_data)"

for i in {1..5}
do
    # deviceshifu return the cleaned data
    deviceshifu_output=$(kubectl exec -it -n deviceshifu nginx -- curl -XPOST -H "Content-Type:application/json" -s deviceshifu-humidity-detector-service.deviceshifu.svc.cluster.local:80/humidity --connect-timeout 5)
    device_output=$(kubectl exec -it -n deviceshifu nginx -- curl -XPOST -H "Content-Type:application/json" -s humidity-detector.devices.svc.cluster.local:11111/humidity --connect-timeout 5)

    device_output_check="$(diff <(echo "$device_output") <(echo "$raw_data") -b)"
    if [[ $device_output == "" ]]
    then
        echo "empty device_output reply"
        exit 1
    elif [[ $device_output_check == "" ]]
    then
        echo "equal device_output reply"
    else
        echo "wrong device_output reply"
        echo "$device_output_check"
        exit 1
    fi
    deviceshifu_output_check="$(diff <(echo "$deviceshifu_output") <(echo "$cleaned_raw_data") -b)"
    if [[ $deviceshifu_output == "" ]]
    then
        echo "empty deviceshifu_output reply"
        exit 1
    elif [[ $deviceshifu_output_check == "" ]]
    then
        echo "equal deviceshifu_output reply"
    else
        echo "wrong deviceshifu_output reply"
        echo "$deviceshifu_output_check"
        exit 1
    fi
done
exit 0
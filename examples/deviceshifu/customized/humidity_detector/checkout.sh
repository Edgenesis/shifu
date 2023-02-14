##!bin/bash
IFS=
cleaned_raw_data=$(cat cleaned_raw_data)
raw_data="$(cat raw_data)"
echo "------"
#echo $cleaned_raw_data
echo "------"
echo $raw_data
echo "------"


for i in {1..5}
do
    # deviceshifu return the cleaned data
    deviceshifu_output=$(kubectl exec -it -n deviceshifu nginx -- curl -H "Content-Type:application/json" -s deviceshifu-humidity-detector-service.deviceshifu.svc.cluster.local:80/humidity --connect-timeout 5)
    device_output=$(kubectl exec -it -n deviceshifu nginx -- curl -XPOST -H "Content-Type:application/json" -s humidity-detector.devices.svc.cluster.local:11111/humidity --connect-timeout 5)

    echo "------"
    #echo $deviceshifu_output
    echo "------"
    echo $device_output
    echo "------"

#    if [[ $device_output == "" || $deviceshifu_output == "" ]]
#    then
#        echo "empty reply"
#        exit 0
#    elif [[ $device_output == $raw_data ]]
    diff <(echo "$device_output") <(echo "$raw_data")
    if [[ "$device_output" == "$raw_data" ]]
    then
        echo "equal reply"
    else
        echo "wrong reply"
        exit 0
    fi
done
exit 1
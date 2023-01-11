curl --header "Content-Type: application/json" \
--request POST --data '{"deviceName":"xyz", "secretName": "test-secret", "serverAddress":"rtsp-server.shifu-app.svc.cluster.local:8554/mystream", "record":true}' \
rtsp-record.shifu-app.svc.cluster.local/register
sleep 5s
curl --header "Content-Type: application/json" \
--request POST --data '{"deviceName":"xyz", "record":false}' \
rtsp-record.shifu-app.svc.cluster.local/update
sleep 1s
curl --header "Content-Type: application/json" \
--request POST --data '{"deviceName":"xyz", "record":true}' \
rtsp-record.shifu-app.svc.cluster.local/update
#sleep 5s
#curl --header "Content-Type: application/json" \
#--request POST --data '{"deviceName":"xyz"}' \
#rtsp-record.shifu-app.svc.cluster.local/unregister

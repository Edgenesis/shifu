docker run -it -p 1883:1883 -v /home/tom/mqtt_test/mosquitto.conf:/mosquitto/config/mosquitto.conf eclipse-mosquitt

mosquitto_sub -h 172.28.15.229 -p 1883 -t /test/test -d

mosquitto_pub -h 172.28.15.229 -d -p 1883 -t /test/test -m "test2333"

kubectl exec -it nginx -- bash

curl edgedevice-led/mqtt_data
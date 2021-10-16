## What can *shifu* liberate the developer from?
TL;DR: *shifu* can save developers at least **45%** effort and time, and cut down at least **80%** maintenance worries about service down.

### Plug and play
You won't need to deal with the mess of configuring new driver, accommodate new devices, learning new API, etc. 

Every time we need to add a new device, we will need to follow a procedure of at least 9 steps:
1. find and install the driver
2. learn the device operations through the driver
3. write new code in the current control system to accommodate the new device, so the system can recognize and operate the device
4. write new code in the current monitor system to accommodate the new device, so we can collect the telemetries of the device 
5. test the new version of the control system with added support of the new device
6. test the new version of the monitor system with added telemetries for the new device
7. deploy the new version of the control system to production environment
8. deploy the new version of the monitor system to production environment
9. finally start using the new device

With *shifu*, we can cut the procedure above from 9 steps to only 5 steps:
1. find and install driver
2. learn the device operations through the driver
3. write the control config of the device
4. write the monitor config of the device
5. start using the new device

As shown above, shifu can save you at least **45%** effort and time.

### High reliability
You will have way less worries about the service down time, as shifu will take care of it.

Most common causes of service down include:
1. software deployment and upgrade
2. software static config change which requires a service reboot
3. software bug
4. bad hardware
5. high load of requests

With *shifu*, we only need to worry about one issue:
1. software bug

Because *shifu* can bring us the following features：
1. service is always available during software deployment/upgrade because of its distributed nature
2. dynamic configs which eliminate the need of service reboot 
3. automatically find good machine to deploy the software if there is a bad hardware
4. an in-house load balancer

As a result, we successfully cut down **80%** issues to worry about - as now we only need to worry about the code of our business logic. 

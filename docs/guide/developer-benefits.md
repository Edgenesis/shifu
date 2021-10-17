## What can *shifu* liberate the developer from?
TL;DR: *shifu* can save developers at least **45%** effort and time, and cut down at least **80%** maintenance worries about service down.

### Plug and play
You won't need to deal with the mess of accommodating new devices. 

Every time we need to add a new device, we will need to follow a procedure of at least 9 steps:
1.  find or implement the the driver
2.  learn the device operations through the driver through the driver and device manual
3.  add new code to control the new device
4.  add new monitoring logic to monitor the device
5.  test the new version of the control system with added support of the new device
6.  test the new version of the monitor system with added telemetries for the new device
7.  stop the current control system
8.  deploy the new version of the control system to production environment
9.  start the upgraded control system
10. stop the current monitor system
11. deploy the new version of the monitor system to production environment
12. start the upgraded monitor system
13. start using the new device

With *shifu*, we can cut the procedure above from 13 steps to only 5 steps:
1. find or implement the the driver
2. learn the device operations through the driver through the driver and device manual
3. write the control and monitor config of the device in a single file
4. start using the new device

As shown above, shifu can save you at least **69%** effort and time.

### High reliability
You will have way less worries about the service down time, as shifu will take care of it.

Most common causes of service down include:
1. software deployment and upgrade
2. driver upgrade
3. software static config change which requires a service reboot
4. device control system bug
5. business logic bug
6. bad hardware
7. high load of requests

With *shifu*, we only need to worry about one issue:
1. business logic bug

Because *shifu* can bring us the following features：
1. service is always available during software and driver deployment/upgrade
2. dynamic configs which eliminate the need of service reboot
3. software will redistribute upon hardware failure
4. an in-house load balancer
5. a kubernetes based device control system

As a result, we successfully cut down **85%** issues to worry about - as now we only need to worry about the code of our business logic. 

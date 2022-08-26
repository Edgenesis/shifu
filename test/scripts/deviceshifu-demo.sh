#!/bin/sh

usage ()
{
  echo "usage: $0 apply/delete edgedevice-[agv/robot-arm/plate-reader/thermometer]"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        kubectl "$1" -f examples/deviceshifu/demo_device/$2
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi

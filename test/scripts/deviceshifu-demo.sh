#!/bin/bash

usage ()
{
  echo "usage: $0 apply/delete edgedevice-[agv/robot-arm/tecan/thermometer]"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        kubectl "$1" -f deviceshifu/examples/demo_device/$2
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi

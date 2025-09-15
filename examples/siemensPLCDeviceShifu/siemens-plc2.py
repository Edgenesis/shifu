import os
import sys
import snap7
import socket
import logging
from flask import Flask, request, jsonify
from typing import Union

# Set up logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

client = snap7.client.Client()

app = Flask(__name__)

# Configuration with environment variables and default values
PLC_ADDRESS = os.environ.get("PLC_ADDRESS")  # No default, must be explicitly set
PLC_PORT = int(os.environ.get("PLC_PORT", "102"))  # Default Siemens S7 port
FLASK_PORT = int(os.environ.get("PLC_CONTAINER_PORT", "11111"))  # Default to 11111 for container
PLC_RACK = int(os.environ.get("PLC_RACK", "0"))  # Default rack number
PLC_SLOT = int(os.environ.get("PLC_SLOT", "1"))  # Default slot number


def edit_single_bit(originalbyte: bytes, digitvalue: int, isset: int) -> bytes:
    if digitvalue > 7:
        digitvalue -= 8
        changebyte = originalbyte[1]
        if isset == 0:
            return bytes([originalbyte[0]]) + bytes([changebyte & ~(1 << digitvalue)])
        else:
            return bytes([originalbyte[0]]) + bytes([changebyte | (1 << digitvalue)])
    else:
        changebyte = originalbyte[0]
        if isset == 0:
            return bytes([changebyte & ~(1 << digitvalue)]) + bytes([originalbyte[1]])
        else:
            return bytes([changebyte | (1 << digitvalue)]) + bytes([originalbyte[1]])


def get_area(rootaddress: str) -> Union[int, None]:
    area_map = {
        "M": snap7.type.Areas.MK,
        "Q": snap7.type.Areas.PA,
        "C": snap7.type.Areas.CT,
        "T": snap7.type.Areas.TM,
    }
    return area_map.get(rootaddress)

def bytearray_to_binary_string(data: bytearray) -> str:
    """Convert a bytearray to a binary string representation."""
    return ' '.join(format(byte, '08b') for byte in data)

@app.route("/sendsinglebit")
def send_single_bit():
    try:
        rootaddress = request.args.get("rootaddress")
        address = int(request.args.get("address", 0))
        start = int(request.args.get("start", 0))
        digit = int(request.args.get("digit", 0))
        value = int(request.args.get("value", 0))

        logger.info(f"Changing single bit: Memory area:{rootaddress}, target address:{address}, setting {digit} digit to {value}")

        area = get_area(rootaddress)
        if area is None:
            return jsonify({"error": f"UnsupportedMemoryArea: root address {rootaddress} not supported!"}), 400

        originalbyte = client.read_area(area, address, start, snap7.type.WordLen.Byte)
        original_binary = bytearray_to_binary_string(originalbyte)
        logger.debug(f"Original value: {original_binary}")

        newbyte = edit_single_bit(originalbyte, digit, value)
        client.write_area(area, address, start, newbyte)

        verifybyte = client.read_area(area, address, start, snap7.type.WordLen.Byte)
        verify_binary = bytearray_to_binary_string(verifybyte)
        logger.info(f"Changed from {original_binary} to {verify_binary}")

        return jsonify({
            "message": f"Changed bit {digit} from {value} to {1-value}",
            "original": original_binary,
            "new": verify_binary
        })

    except Exception as e:
        logger.error(f"Error in send_single_bit: {str(e)}")
        return jsonify({"error": str(e)}), 500


@app.route("/getcontent")
def get_content():
    try:
        rootaddress = request.args.get("rootaddress")
        address = int(request.args.get("address", 0))
        start = int(request.args.get("start", 0))

        logger.info(
            f"Getting content from memory areas: Reading from {rootaddress} for address {address} starting {start}"
        )

        area = get_area(rootaddress)
        if area is None:
            return (
                jsonify(
                    {
                        "error": f"UnsupportedMemoryArea: root address {rootaddress} not supported!"
                    }
                ),
                400,
            )

        res = client.read_area(area, address, start, snap7.type.WordLen.Byte)
        resbin = bin(int.from_bytes(res, sys.byteorder))

        if len(resbin) < 18:
            resbin = (
                resbin.split("b")[0]
                + "b"
                + (18 - len(resbin)) * "0"
                + resbin.split("b")[1]
            )

        logger.debug(f"Got {resbin}, length {len(resbin)}")
        return jsonify({"result": resbin})

    except Exception as e:
        logger.error(f"Error in get_content: {str(e)}")
        return jsonify({"error": str(e)}), 500


@app.route("/getcpuordercode")
def get_cpu_ordercode():
    try:
        logger.info("Getting CPU order code...")
        order_code = client.get_order_code()
        
        # Decode the bytes object to a string
        order_code_str = order_code.OrderCode.decode('utf-8').strip()
        
        logger.info(f"CPU order code is {order_code_str}")
        return jsonify({"order_code": order_code_str})
    except Exception as e:
        logger.error(f"Error in get_cpu_ordercode: {str(e)}")
        return jsonify({"error": str(e)}), 500


def get_ip(d: str) -> Union[str, bool]:
    # If it's already an IP address, return it
    try:
        socket.inet_aton(d)
        return d
    except socket.error:
        pass

    # If not an IP, then resolve
    try:
        data = socket.gethostbyname(d)
        logger.info(f"Resolved IP: {data}")
        return data
    except Exception as e:
        logger.error(f"Failed to resolve IP for {d}: {str(e)}")
        return False


if __name__ == "__main__":
    if not PLC_ADDRESS:
        logger.error("PLC_ADDRESS environment variable is not set. Exiting.")
        sys.exit(1)

    try:
        ip = get_ip(PLC_ADDRESS)
        if not ip:
            raise ValueError(f"Could not resolve IP for {PLC_ADDRESS}")

        client.connect(ip, PLC_RACK, PLC_SLOT, PLC_PORT)
        logger.info(f"Connected to PLC at {ip}:{PLC_PORT}")

        app.run(host="0.0.0.0", port=FLASK_PORT)
    except Exception as e:
        logger.error(f"Failed to start the application: {str(e)}")
        sys.exit(1)
    finally:
        if client.get_connected():
            client.disconnect()
            logger.info("Disconnected from PLC")

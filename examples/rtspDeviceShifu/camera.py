import cv2, logging, os, requests, time
from flask import Flask, Response
from requests.auth import HTTPDigestAuth, HTTPBasicAuth
from threading import Thread

app = Flask(__name__)

ip = os.environ.get("IP_CAMERA_ADDRESS")
CAMERA_USERNAME = os.environ.get("IP_CAMERA_USERNAME")
CAMERA_PASSWORD = os.environ.get("IP_CAMERA_PASSWORD")
port = os.environ.get("IP_CAMERA_CONTAINER_PORT")

CAMERA_CTRL_MOVE_UP = '<?xml version="1.0" encoding="UTF-8"?><PTZData><pan>0</pan><tilt>60</tilt></PTZData>'
CAMERA_CTRL_MOVE_DOWN = '<?xml version="1.0" encoding="UTF-8"?><PTZData><pan>0</pan><tilt>-60</tilt></PTZData>'
CAMERA_CTRL_MOVE_LEFT = '<?xml version="1.0" encoding="UTF-8"?><PTZData><pan>-60</pan><tilt>0</tilt></PTZData>'
CAMERA_CTRL_MOVE_RIGHT = '<?xml version="1.0" encoding="UTF-8"?><PTZData><pan>60</pan><tilt>0</tilt></PTZData>'
CAMERA_CTRL_MOVE_STOP = '<?xml version="1.0" encoding="UTF-8"?><PTZData><pan>0</pan><tilt>0</tilt></PTZData>'

CAMERA_CTRL_MOVE_DICT = {
    "up": CAMERA_CTRL_MOVE_UP,
    "down": CAMERA_CTRL_MOVE_DOWN,
    "left": CAMERA_CTRL_MOVE_LEFT,
    "right": CAMERA_CTRL_MOVE_RIGHT,
}


class VideoGet:
    """
    Class that continuously gets frames from a VideoCapture object
    with a dedicated thread.
    """

    def __init__(self, ip, username, password):
        self.stream = cv2.VideoCapture("rtsp://{}:{}@{}".format(CAMERA_USERNAME, CAMERA_PASSWORD, ip))
        (self.grabbed, self.frame) = self.stream.read()
        self.stopped = False

    def start(self):
        Thread(target=self.get, args=()).start()
        return self

    def get(self):
        while not self.stopped:
            if not self.grabbed:
                self.restart()
            else:
                (self.grabbed, self.frame) = self.stream.read()

    def stop(self):
        self.stopped = True

    def restart(self):
        print("capture failed, restarting...")
        self.stream.release()
        print("released stream and allocate new one...")
        try:
            self.stream = cv2.VideoCapture("rtsp://{}:{}@{}".format(CAMERA_USERNAME, CAMERA_PASSWORD, ip))
            print("capture reopen success!")
        except Exception as e:
            print("error open stream, error: {}".format(e))
        if not self.stream.isOpened():
            print("stream is not opened, try opening...")
            self.stream.open("rtsp://{}:{}@{}".format(CAMERA_USERNAME, CAMERA_PASSWORD, ip))
        # time.sleep(600)



@app.route('/capture')
def capture():
    try:    
        if not video_getter.stopped:
            ret, frame = video_getter.grabbed, video_getter.frame
            if ret:
                retval, buffer = cv2.imencode('.jpg', frame)
                byte_frame = buffer.tobytes()
                print("Image captured!")
                return Response(byte_frame, mimetype='image/jpeg')
            else:
                print("cannot capture frame from cv2")
                return "cannot capture frame from cv2\n", 400
    except Exception as e:
        print("error capture picture, error: {}".format(e))
        # return False
        return "cannot capture frame from cv2\n", 400


def stream(ip, username, password):
    try:
        print("start streaming!")
        while True:
            success, frame = video_getter.grabbed, video_getter.frame
            if not success:
                break
            else:
                reducedframe = cv2.resize(frame, (0,0), fx=0.5, fy=0.5) 
                ret, buffer = cv2.imencode('.jpeg', reducedframe)
                framedata = buffer.tobytes()
                yield (b'--frame\r\n'
                    b'Content-Type: image/jpeg\r\n\r\n' + framedata + b'\r\n')  # concat frame one by one and show result
    except Exception as e:
        print("error capture picture, error: {}".format(e))
        return False


def getCameraInfoWithAuth(s, ip, auth):
    result = None
    s.auth = auth
    try:
        r = s.get('http://' + ip + '/PSIA/System/deviceInfo')
        if r.ok:
            result = r.content
        else:
            r = s.get('http://' + ip + '/ISAPI/System/deviceInfo')
            if r.ok:
                result = r.content
            else:
                print("{} failed".format(type(auth)))
    except Exception as e:
        result = None
        print("error trying {}, {}".format(type(auth), e))
    
    return result


def moveCameraWithAuth(s, ip, auth, direction):
    result = None
    s.auth = auth
    try:
        headers = {'Content-Type': 'application/xml'}
        r = s.put('http://' + ip + '/ISAPI/PTZCtrl/channels/1/continuous', data=CAMERA_CTRL_MOVE_DICT[direction], headers=headers)
        if r.ok:
            time.sleep(0.2)
            r = s.put('http://' + ip + '/ISAPI/PTZCtrl/channels/1/continuous', data=CAMERA_CTRL_MOVE_STOP, headers=headers)
            result = r.content
        else:
            print("{} failed, message: {}".format(type(auth), r.content))
    except Exception as e:
        result = None
        print("error trying {}, {}".format(type(auth), e))
    
    return result


def moveCamera(direction):
    with requests.Session() as s:
        result = None
        print("try HTTPDigestAuth")
        auth = HTTPDigestAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
        result = moveCameraWithAuth(s, ip, auth, direction)

        if result is None:
            print("try HTTPBasicAuth")
            auth = HTTPBasicAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
            result = moveCameraWithAuth(s, ip, auth, direction)
            if result is None:
                print("all authentication failed for device")
                return False

        return True


@app.route('/info')
def getCameraInfo():
    with requests.Session() as s:
        result = None
        print("try HTTPDigestAuth")
        auth = HTTPDigestAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
        result = getCameraInfoWithAuth(s, ip, auth)

        if result is None:
            print("try HTTPBasicAuth")
            auth = HTTPBasicAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
            result = getCameraInfoWithAuth(s, ip, auth)

        if result is None:
            print("all authentication failed for device")
            return "all authentication failed for device\n", 400

        return Response(result, mimetype='text/xml')


@app.route('/stream')
def video_feed():
    #Video streaming route. Put this in the src attribute of an img tag
    return Response(stream(ip, CAMERA_USERNAME, CAMERA_PASSWORD), mimetype='multipart/x-mixed-replace; boundary=frame')


@app.route('/move/<direction>')
def move_camera(direction=None):
    print(CAMERA_CTRL_MOVE_DICT.keys())
    print("direction is {}".format(direction))
    if direction is None or direction not in CAMERA_CTRL_MOVE_DICT.keys():
        return 'Please specify move direction, /move/(up/down/left/right)\n', 400
    if moveCamera(direction):
        return 'Success', 200
    else:
        return 'cannot move camera',400


if __name__ == "__main__":
    video_getter = VideoGet(ip, CAMERA_USERNAME, CAMERA_PASSWORD).start()
    app.run(host="0.0.0.0", port=port)

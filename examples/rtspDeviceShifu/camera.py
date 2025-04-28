import cv2, logging, os, requests, time, threading, queue, gc
from flask import Flask, Response, jsonify
from requests.auth import HTTPDigestAuth, HTTPBasicAuth
from threading import Thread

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Camera settings
CAMERA_WIDTH = 640
CAMERA_HEIGHT = 360
CAMERA_FPS = 15
JPEG_QUALITY = 80
BUFFER_SIZE = 1
MAX_CLIENTS = 5

# Environment variables
ip = os.environ.get("IP_CAMERA_ADDRESS")
http_port = os.environ.get("IP_CAMERA_HTTP_PORT")
rtsp_port = os.environ.get("IP_CAMERA_RTSP_PORT")
CAMERA_USERNAME = os.environ.get("IP_CAMERA_USERNAME")
CAMERA_PASSWORD = os.environ.get("IP_CAMERA_PASSWORD")
port = os.environ.get("IP_CAMERA_CONTAINER_PORT")

# Camera control constants
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
    def __init__(self, ip, rtsp_port, username, password):
        self.ip = ip
        self.rtsp_port = rtsp_port
        self.username = username
        self.password = password
        self.stream = None
        self.frame_queue = queue.Queue(maxsize=BUFFER_SIZE)
        self._should_stop = False
        self._capture_thread = None
        self._last_frame = None
        self._client_count = 0
        self._last_gc_time = time.time()
        self._reconnect_attempts = 0
        self._max_reconnect_attempts = 5
        self._reconnect_delay = 1  # seconds
        self.start()

    def _reconnect(self):
        if self.stream is not None:
            self.stream.release()
        
        rtsp_url = f"rtsp://{self.username}:{self.password}@{self.ip}{self.rtsp_port}"
        logger.info(f"Attempting to reconnect to RTSP stream: {rtsp_url}")
        
        self.stream = cv2.VideoCapture(rtsp_url)
        if not self.stream.isOpened():
            logger.error("Failed to reconnect to RTSP stream")
            return False
        logger.info("Successfully reconnected to RTSP stream")
        return True

    def _capture_frames(self):
        frame_count = 0
        last_fps_time = time.time()
        
        while not self._should_stop:
            try:
                if self._client_count == 0:
                    time.sleep(0.1)
                    continue

                if self.stream is None or not self.stream.isOpened():
                    if self._reconnect_attempts < self._max_reconnect_attempts:
                        if self._reconnect():
                            self._reconnect_attempts = 0
                        else:
                            self._reconnect_attempts += 1
                            time.sleep(self._reconnect_delay)
                            continue
                    else:
                        logger.error("Max reconnection attempts reached")
                        time.sleep(1)
                        continue

                success, frame = self.stream.read()
                if not success:
                    logger.warning("Failed to read frame, attempting reconnection")
                    if not self._reconnect():
                        time.sleep(self._reconnect_delay)
                        continue
                    continue

                if frame.shape[1] != CAMERA_WIDTH or frame.shape[0] != CAMERA_HEIGHT:
                    frame = cv2.resize(frame, (CAMERA_WIDTH, CAMERA_HEIGHT), 
                                    interpolation=cv2.INTER_LINEAR)

                self._last_frame = frame.copy()

                try:
                    if self.frame_queue.full():
                        self.frame_queue.get_nowait()
                except queue.Empty:
                    pass

                self.frame_queue.put(frame, timeout=0.1)

                frame_count += 1
                if frame_count >= 30:
                    current_time = time.time()
                    fps = frame_count / (current_time - last_fps_time)
                    logger.info(f"Capture FPS: {fps:.2f}")
                    frame_count = 0
                    last_fps_time = current_time

                current_time = time.time()
                if current_time - self._last_gc_time > 60:
                    gc.collect()
                    self._last_gc_time = current_time

            except Exception as e:
                logger.error(f"Error in capture thread: {str(e)}")
                time.sleep(0.1)

    def start(self):
        self._should_stop = False
        if self.stream is None:
            self._reconnect()
        self._capture_thread = Thread(target=self._capture_frames)
        self._capture_thread.daemon = True
        self._capture_thread.start()
        return self

    def get_frame(self):
        try:
            frame = self.frame_queue.get(timeout=0.1)
            if frame is None and self._last_frame is not None:
                frame = self._last_frame.copy()
            
            if frame is not None:
                encode_param = [
                    int(cv2.IMWRITE_JPEG_QUALITY), JPEG_QUALITY,
                    int(cv2.IMWRITE_JPEG_OPTIMIZE), 1,
                ]
                ret, buffer = cv2.imencode(".jpg", frame, encode_param)
                return buffer.tobytes()
            return None
        except queue.Empty:
            if self._last_frame is not None:
                encode_param = [
                    int(cv2.IMWRITE_JPEG_QUALITY), JPEG_QUALITY,
                    int(cv2.IMWRITE_JPEG_OPTIMIZE), 1,
                ]
                ret, buffer = cv2.imencode(".jpg", self._last_frame, encode_param)
                return buffer.tobytes()
            return None
        except Exception as e:
            logger.error(f"Error getting frame: {str(e)}")
            return None

    def stop(self):
        self._should_stop = True
        if self._capture_thread:
            self._capture_thread.join(timeout=1.0)
            self._capture_thread = None
        self.stream.release()
        gc.collect()

    def add_client(self):
        self._client_count += 1

    def remove_client(self):
        self._client_count -= 1

def getCameraInfoWithAuth(s, ip, http_port, auth):
    result = None
    s.auth = auth
    try:
        r = s.get('http://' + ip + http_port + '/PSIA/System/deviceInfo')
        if r.ok:
            result = r.content
        else:
            r = s.get('http://' + ip + http_port + '/ISAPI/System/deviceInfo')
            if r.ok:
                result = r.content
            else:
                logger.warning(f"{type(auth)} failed")
    except Exception as e:
        result = None
        logger.error(f"Error trying {type(auth)}, {e}")
    
    return result

def moveCameraWithAuth(s, ip, http_port, auth, direction):
    result = None
    s.auth = auth
    try:
        getCameraInfoWithAuth(s, ip, http_port, auth)
        headers = {'Content-Type': 'application/xml'}
        r = s.put('http://' + ip + http_port + '/ISAPI/PTZCtrl/channels/1/continuous', 
                 data=CAMERA_CTRL_MOVE_DICT[direction], headers=headers)
        if r.ok:
            time.sleep(0.2)
            r = s.put('http://' + ip + http_port + '/ISAPI/PTZCtrl/channels/1/continuous', 
                     data=CAMERA_CTRL_MOVE_STOP, headers=headers)
            result = r.content
        else:
            logger.warning(f"{type(auth)} failed, message: {r.content}")
    except Exception as e:
        result = None
        logger.error(f"Error trying {type(auth)}, {e}")
    
    return result

def moveCamera(direction):
    with requests.Session() as s:
        result = None
        logger.info("Trying HTTPDigestAuth")
        auth = HTTPDigestAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
        result = moveCameraWithAuth(s, ip, http_port, auth, direction)

        if result is None:
            logger.info("Trying HTTPBasicAuth")
            auth = HTTPBasicAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
            result = moveCameraWithAuth(s, ip, http_port, auth, direction)
            if result is None:
                logger.error("All authentication failed for device")
                return False

        return True

@app.route('/info')
def getCameraInfo():
    with requests.Session() as s:
        result = None
        logger.info("Trying HTTPDigestAuth")
        auth = HTTPDigestAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
        result = getCameraInfoWithAuth(s, ip, http_port, auth)

        if result is None:
            logger.info("Trying HTTPBasicAuth")
            auth = HTTPBasicAuth(CAMERA_USERNAME, CAMERA_PASSWORD)
            result = getCameraInfoWithAuth(s, ip, http_port, auth)

        if result is None:
            logger.error("All authentication failed for device")
            return jsonify({"error": "All authentication failed for device"}), 400

        return Response(result, mimetype='text/xml')

@app.route('/capture')
def capture():
    try:
        if not video_getter._capture_thread:
            video_getter.start()
        
        frame_data = video_getter.get_frame()
        if frame_data is not None:
            return Response(frame_data, mimetype='image/jpeg')
        else:
            logger.error("Unable to capture frame")
            return jsonify({"error": "Unable to capture frame"}), 500
    except Exception as e:
        logger.error(f"Capture error: {str(e)}")
        return jsonify({"error": str(e)}), 500

@app.route('/stream')
def video_feed():
    def generate():
        try:
            video_getter.add_client()
            while True:
                frame_data = video_getter.get_frame()
                if frame_data is None:
                    time.sleep(0.01)
                    continue

                yield (
                    b"--frame\r\n"
                    b"Content-Type: image/jpeg\r\n\r\n" + frame_data + b"\r\n"
                )
        finally:
            video_getter.remove_client()

    return Response(generate(), mimetype='multipart/x-mixed-replace; boundary=frame')

@app.route('/move/<direction>')
def move_camera(direction=None):
    if direction is None or direction not in CAMERA_CTRL_MOVE_DICT.keys():
        return jsonify({"error": "Please specify move direction (up/down/left/right)"}), 400
    
    if moveCamera(direction):
        return jsonify({"message": "Success"}), 200
    else:
        return jsonify({"error": "Cannot move camera"}), 400

if __name__ == "__main__":
    if http_port:
        http_port = ':' + http_port

    if rtsp_port:
        rtsp_port = ':' + rtsp_port

    video_getter = VideoGet(ip, rtsp_port, CAMERA_USERNAME, CAMERA_PASSWORD)
    app.run(host="0.0.0.0", port=port)

from flask import Flask, request, make_response

app = Flask(__name__)

@app.route('/get_file')
def get_file():
    response = make_response("file_content", 200)
    response.mimetype = "text/plain"
    return response

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=12345)

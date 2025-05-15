from flask import Flask
import logging

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

app = Flask(__name__)

@app.route('/')
def hello():
    logger.debug("Got request!")
    return 'Hello World!'

if __name__ == '__main__':
    logger.info("Starting minimal test server...")
    app.run(host='0.0.0.0', port=8089, debug=True) 
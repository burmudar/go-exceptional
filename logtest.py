import logging
import time
from logging.handlers import RotatingFileHandler

def main():
    logger = logging.getLogger("Rotating Log")
    logger.setLevel(logging.DEBUG)

    handler = RotatingFileHandler("test.log", maxBytes=20, backupCount=5)
    logger.addHandler(handler)

    for i in range(500):
        logger.debug("Test {}".format(i))
        time.sleep(1.5)


if __name__ == "__main__":
    main()

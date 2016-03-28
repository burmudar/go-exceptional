import logging
import time
import datetime
from logging.handlers import RotatingFileHandler

def main():
    logger = logging.getLogger("Rotating Log")
    logger.setLevel(logging.DEBUG)

    handler = RotatingFileHandler("test.log", maxBytes=200000000000000, backupCount=5)
    logger.addHandler(handler)

    days = 0
    now = datetime.datetime.now() - datetime.timedelta(days=2)
    for i in range(500):
        timestamp = now + datetime.timedelta(days=days)
        if (i + 1) % 10 == 0:
            days = days + 1
        logErrorLine(timestamp, logger)
        time.sleep(1.5)

def logErrorLine(timestamp, logger):
    line = "[{}] ERROR client.AirtelService:54 - 0833574730 : Encountered an error while querying balance : TranRef[testRef]".format(timestamp.strftime("%Y-%m-%d %H:%M:%S,000"))
    logger.debug(line)
    time.sleep(1.5)
    line = "Caused by: com.mysql.jdbc.exceptions.jdbc4.MySQLSyntaxErrorException: UPDATE command denied to user 'fsi_app'@'10.0.1.231' for table 'recharge_provider_setting'"
    logger.debug(line)
    time.sleep(1.5)

if __name__ == "__main__":
    main()

import structlog
import logging
import logging.config
import sys


def drop_record(_, __, event_dict):
    event_dict.pop('_record', None)
    event_dict.pop('_from_structlog', None)
    return event_dict


PROCESSORS = [
    structlog.stdlib.ExtraAdder(),
    structlog.stdlib.add_logger_name,
    structlog.stdlib.add_log_level,
    structlog.stdlib.PositionalArgumentsFormatter(),
    structlog.processors.TimeStamper(fmt="iso"),
    structlog.processors.StackInfoRenderer(),
    structlog.processors.dict_tracebacks,
    structlog.processors.UnicodeDecoder(),
    structlog.processors.EventRenamer("message"),
    drop_record,
    structlog.processors.JSONRenderer(),
]

CONFIG = {
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {
        "json": {
            "()": structlog.stdlib.ProcessorFormatter,
            "processors": PROCESSORS,
        },
    },
    "handlers": {
        "default": {
            "level": "DEBUG",
            "class": "logging.StreamHandler",
            "stream": sys.stdout,
            "formatter": "json",
        },
    },
    "loggers": {
        "": {
            "handlers": ["default"],
            "level": "INFO",
            "propagate": True,
        },
    },
}


logging.config.dictConfig(CONFIG)

structlog.configure(
    processors=PROCESSORS,
    logger_factory=structlog.stdlib.LoggerFactory(),
    wrapper_class=structlog.stdlib.BoundLogger,
    cache_logger_on_first_use=True,
)


def get_logger(name: str) -> logging.Logger:
    return logging.getLogger(name)

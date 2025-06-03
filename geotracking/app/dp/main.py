import argparse
import os
import signal
import sys

import init
import log
import sync


LOG = log.get_logger(__name__)


parser = argparse.ArgumentParser(
    prog="dp",
    description="ClickHouse data integration tool",
)

parser.add_argument(
    "-c",
    dest="dsn",
    default=os.getenv("CLICKHOUSE_DSN"),
    help="ClickHouse connection URL",
)

subparsers = parser.add_subparsers(
    dest="command",
    help="available commands",
    required=True,
)

init_args = subparsers.add_parser(
    name="init",
    help="Initialize tables",
)

sync_args = subparsers.add_parser(
    name="sync",
    help="Run tables synchromization",
)
sync_args.add_argument(
    "-i",
    type=int,
    dest="interval",
    default=os.getenv("SYNC_INTERVAL", "60"),
    help="Sync interval",
)

def main() -> None:
    args = parser.parse_args()
    if args.command == "init":
        init.run_init(args.dsn)
        return
    if args.command == "sync":

        def on_stop(signum, _):
            LOG.warning("terminating due to signal", extra={"signum": signum})
            sync.RUNNING = False

        signal.signal(signal.SIGTERM, on_stop)
        sync.run_sync(args.dsn, args.interval)
        return
    sys.stderr.write(f"unpexpected command {args.command}")
    sys.exit(2)


if __name__ == "__main__":
    main()

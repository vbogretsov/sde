import time
from dataclasses import dataclass
from datetime import datetime
from datetime import timezone
from datetime import timedelta
from concurrent.futures import ThreadPoolExecutor

from clickhouse_driver import connect
from clickhouse_driver.dbapi.connection import Connection
from clickhouse_driver.dbapi.cursor import Cursor
from clickhouse_driver import Client

import log


LOG = log.get_logger(__name__)


PATH_FMT = "{root}/year={year}/month={month:02}/day={day:02}/hour={hour:02}/*.parquet"

QUERY_FMT = """
INSERT INTO {table} SELECT * FROM s3Cluster(cluster, '{path}')
    WHERE {order} >= {watermark}
"""


RUNNING = True


@dataclass
class Task:
    table: str
    order: str
    root: str

    def format_query(self, watermark: int) -> list[str]:
        queries: list[str] = []
        if watermark > 0:
            ts_from = datetime.fromtimestamp(watermark / 1000, tz=timezone.utc)
            ts_to = datetime.now(tz=timezone.utc)
            ts = ts_from.replace(minute=0, second=0, microsecond=0)

            while ts <= ts_to:
                path = PATH_FMT.format(
                    root=self.root,
                    year=ts.year,
                    month=ts.month,
                    day=ts.day,
                    hour=ts.hour,
                )
                query = QUERY_FMT.format(
                    table=self.table,
                    order=self.order,
                    path=path,
                    watermark=watermark,
                )
                queries.append(query)
                ts += timedelta(hours=1)
        else:
            path = f"{self.root}/year=*/month=*/day=*/hour=*/*.parquet"
            query = f"INSERT INTO {self.table} SELECT * FROM s3Cluster(cluster, '{path}')"
            queries.append(query)

        return queries


TASKS = [
    Task(
        table="app.track_locations",
        order="__ts_ms",
        root="http://s3:9000/data/topics/app.app.track_locations",
    ),
    Task(
        table="app.track_routes",
        order="__ts_ms",
        root="http://s3:9000/data/topics/app.app.track_routes",
    ),
]


def run_sync(dsn: str, interval: int) -> None:
    LOG.info("starting synchronization")
    with ThreadPoolExecutor() as pool:
        while RUNNING:
            LOG.info("running sync")
            results = pool.map(lambda t: _do_sync(dsn, t), TASKS)
            _  = list(results)
            time.sleep(interval)


def _do_sync(dsn: str, task: Task) -> None:
    client = Client.from_url(dsn)
    _process_table(client, task)


def _process_table(client: Client, task: Task) -> None:
    try:
        logctx = {"table": task.table}
        LOG.info("processing table", extra=logctx)
        wm_result = client.execute(f"select max({task.order}) from {task.table}")
        if not wm_result:
            watermark = 0
        else:
            watermark = wm_result[0][0]
        logctx["watermark"] = watermark
        LOG.info("synchronizing data from", extra=logctx)
        queries = task.format_query(watermark)
        for query in queries:
            nrows = client.execute(query)
            LOG.info("executed query", extra={
                **logctx,
                "query": query,
                "nrows": nrows,
            })
        LOG.info("table has been processed", extra=logctx)
    except:
        LOG.exception(f"failed to process table {task.table}")

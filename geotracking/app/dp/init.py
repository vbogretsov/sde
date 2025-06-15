from clickhouse_driver import connect
from clickhouse_driver.dbapi.cursor import Cursor

import log

LOG = log.get_logger(__name__)


MON_LOGS = """
CREATE TABLE mon.logs
ON CLUSTER cluster
(
    `time`      DateTime64(9, 'UTC'),
    `tag`       String,
    `log`       String,
    `partnum`   UInt32 MATERIALIZED toUInt32(formatDateTime(time, '%Y%m%d%H'))
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/mon/logs/{shard}', '{replica}')
PARTITION BY partnum
ORDER BY (time, tag)
SETTINGS storage_policy = 's3_main';
"""

# TRACK_LOCATIONS = """
# CREATE TABLE app.raw_track_locations
# ON CLUSTER cluster
# (
#    `_id`            String,
#    `_cid`           String,
#    `updated_at`     DateTime64(3, 'UTC'),
#    `lat`            Float64,
#    `lng`            Float64,
#    `__op`           String,
#    `__ts_ms`        Int64,
# )
# ENGINE = S3(s3_track_locations)
# PARTITION BY toUInt32(formatDateTime(updated_at, '%Y%m%d%H'))
# SETTINGS
#     filesystem_cache_name = 'cache_for_s3',
#     enable_filesystem_cache = 1;
# """

TRACK_LOCATIONS = """
CREATE TABLE app.track_locations
ON CLUSTER cluster
(
   `_id`            String,
   `_cid`           String,
   `updated_at`     DateTime64(3, 'UTC'),
   `lat`            Float64,
   `lng`            Float64,
   `h3_1`           Int64,
   `h3_2`           Int64,
   `h3_3`           Int64,
   `h3_4`           Int64,
   `h3_5`           Int64,
   `h3_6`           Int64,
   `h3_7`           Int64,
   `h3_8`           Int64,
   `h3_9`           Int64,
   `__op`           String,
   `__ts_ms`        Int64
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/track_locations/{shard}', '{replica}')
PARTITION BY toYYYYMMDD(updated_at)
ORDER BY (_id, __ts_ms)
SETTINGS storage_policy = 's3_main';
"""

TRACK_ROUTES = """
CREATE TABLE app.track_routes
ON CLUSTER cluster
(
   `_id`            String,
   `_cid`           String,
   `points`         Array(Array(Nullable(Float64))),
   `updated_at`     DateTime64(3, 'UTC'),
   `__op`           String,
   `__ts_ms`        Int64
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/track_routes/{shard}', '{replica}')
PARTITION BY toYYYYMMDD(updated_at)
ORDER BY (_id, __ts_ms)
SETTINGS storage_policy = 's3_main';
"""

TRACK_LOCATIONS2 = """
CREATE TABLE app.track_locations2
ON CLUSTER cluster
(
   `_id`            String,
   `_cid`           String,
   `updated_at`     DateTime64(3, 'UTC'),
   `lat`            Float64,
   `lng`            Float64,
   `h3_1`           Int64,
   `h3_2`           Int64,
   `h3_3`           Int64,
   `h3_4`           Int64,
   `h3_5`           Int64,
   `h3_6`           Int64,
   `h3_7`           Int64,
   `h3_8`           Int64,
   `h3_9`           Int64,
   `__op`           String,
   `__ts_ms`        Int64
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/track_locations2/{shard}', '{replica}')
PARTITION BY toYYYYMMDD(updated_at)
ORDER BY (_id, __ts_ms)
SETTINGS storage_policy = 's3_cache';
"""

TRACK_ROUTES2 = """
CREATE TABLE app.track_routes2
ON CLUSTER cluster
(
   `_id`            String,
   `_cid`           String,
   `points`         Array(Array(Nullable(Float64))),
   `updated_at`     DateTime64(3, 'UTC'),
   `__op`           String,
   `__ts_ms`        Int64
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/track_routes2/{shard}', '{replica}')
PARTITION BY toYYYYMMDD(updated_at)
ORDER BY (_id, __ts_ms)
SETTINGS storage_policy = 's3_cache';
"""


def run_init(dsn: str) -> None:
    LOG.info("initializing ClickHouse")
    conn = connect(dsn)
    try:
        with conn.cursor() as cursor:
            _create_table(cursor, "mon.logs", MON_LOGS)
            _create_table(cursor, "app.track_locations", TRACK_LOCATIONS)
            _create_table(cursor, "app.track_routes", TRACK_ROUTES)
            # _create_table(cursor, "app.track_locations2", TRACK_LOCATIONS2)
            # _create_table(cursor, "app.track_routes2", TRACK_ROUTES2)
    finally:
        conn.close()


def _create_table(cursor: Cursor, name: str, sql: str) -> None:
    try:
        LOG.info("creating table", extra={"table": name})
        cursor.execute(sql)
    except:
        LOG.exception("failed to create table")

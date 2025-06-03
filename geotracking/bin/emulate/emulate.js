const OSRM = require('./node_modules/@project-osrm/osrm/lib/binding_napi_v8/node_osrm.node').OSRM;

const DATAPATH = '../../data/osrm/central-fed-district-latest.osrm';
const AREA = [[37.3688, 55.5692], [37.8557, 55.9119]]; // Moscow
// const AREA = [[37.486715, 55.735814], [37.486715, 55.778334]] // Moscow Center

const BASE_URL = 'http://localhost:8000';
const LOCATIONS_URL = `${BASE_URL}/api/track/locations`;
const ROUTES_URL = `${BASE_URL}/api/track/routes`;

let countPostLocations = 0;
let countPostRoutes = 0;

const osrm = new OSRM({
  path: DATAPATH,
  algorithm: 'MLD',
});

function random(a, b) {
  return a + Math.random() * (b - a);
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function postRoute(user) {
  const req = {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      uid: user.name,
      route: user.route,
    }),
  }

  const res = await fetch(ROUTES_URL, req);
  countPostRoutes += 1;

  if (!res.ok) {
    const err = await res.text();
    console.error(`POST route failed`, { req, err });
  } else {
    // console.log(`route sent user=${user.name} size=${user.route.length}`);
  }
}

async function postLocation(user) {
  const lat = user.route[user.step][0];
  const lon = user.route[user.step][1];

  const req = {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      uid: user.name,
      loc: [lat, lon],
    }),
  };

  const res = await fetch(LOCATIONS_URL, req);
  countPostLocations += 1;

  if (!res.ok) {
    const err = await res.text();
    console.error(`POST location failed`, { req, err });
  }

  // await sleep(1000);
}

function randomPoint(area) {
  return [
    random(area[0][0], area[1][0]),
    random(area[0][1], area[1][1]),
  ]
}

function lastItem(array) {
  return array[array.length - 1];
}

function generateTask(user) {
  return [
    user.route.length === 0 ? randomPoint(AREA) : lastItem(user.route),
    randomPoint(AREA),
  ]
}

function generateRoute(user) {
  return new Promise((resolve, reject) => {
    osrm.route(
      {
        coordinates: generateTask(user),
        steps: true,
        generate_hints: false,
        overview: 'full',
        geometries: 'geojson',
        skip_waypoints: true,
      },
      (err, result) => {
        if (err) {
          reject(err);
        } else {
          resolve(result.routes[0].geometry.coordinates.map(c => [c[1], c[0]]));
        }
      }
    );
  })
}

async function iter(users) {
  const user = users[Math.floor(random(0, users.length))];
  if (user.locked) {
    return;
  }

  user.locked = true;
  try {
    if (user.step >= user.route.length) {
      user.route = await generateRoute(user);
      user.step = 0;
      await postRoute(user);
    }
    await postLocation(user);
    user.step += 1;
  } finally {
    user.locked = false;
  }
}

async function main() {
  console.log(process.env.DYLD_LIBRARY_PATH);

  const numUsers = parseInt(process.env.NUM_USERS || '1');
  const minUserId = parseInt(process.env.MIN_USER_ID || '1');
  const concurrency = parseInt(process.env.CONCURRENCY || '1');
  console.debug(`Emulating ${numUsers} users with ${concurrency} concurrency minId=${minUserId}`);

  const users = Array.from({ length: numUsers }, (_, i) => ({
    name: `user-${String(minUserId + i).padStart(Math.floor(Math.log10(minUserId + numUsers)), '0')}`,
    step: 0,
    route: [],
    locked: false,
  }));

  const startedAt = new Date();

  function stop(signal) {
    const stoppedAt = new Date();
    const duration = Math.floor((stoppedAt - startedAt) / 1000)

    console.log(`\nTerminating due to signal ${signal}`);
    console.log(`
  Duration            ${duration}s

  POST locations:     ${countPostLocations/duration} req/seq
  POST routes:        ${countPostRoutes/duration}    req/seq
    `);

    process.exit(0);
  }

  process.on('SIGTERM', stop);
  process.on('SIGINT', stop);

  while (true) {
    const promises = [];
    for (let i = 0; i < concurrency; i++) {
      promises.push(iter(users));
    }
    await Promise.all(promises);
  }
}

(async () => {
  await main();
})();

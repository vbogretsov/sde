import React, { useCallback, useState, useRef, useEffect, useMemo } from "react";
import { BrowserRouter as Router, Route, Routes, useParams } from "react-router-dom";
import "./App.css";

import Map, { Marker, Source, Layer } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import geohash from "ngeohash";

function debounce(func: (...args: any[]) => void, delay: number) {
  let timeout: NodeJS.Timeout;
  return (...args: any[]) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), delay);
  };
}

const points = [
  { longitude: 27.5590, latitude: 53.9045, label: 180 }, // Minsk
  { longitude: 37.6173, latitude: 55.7558, label: 230 }, // Moscow
  { longitude: 39.7015, latitude: 47.2357, label: 150 }, // Rostov
  { longitude: 24.7536, latitude: 59.4370, label: 120 }, // Tallinn
];

const lineCoordinates = [
  [27.5590, 53.9045], // Minsk
  [39.7015, 47.2357], // Rostov
];

function adjustZommLevelV2(zoom: number): number {
  if (zoom < 4) {
    return 4; // Minimum zoom level
  }
  if (zoom > 20) {
    return 20; // Maximum zoom level
  }
  return Math.round(zoom); // Round to the nearest integer
}

function adjustZommLevelV3(zoom: number): number {
  if (zoom > 15) {
    return 15; // Maximum zoom level
  }
  const v = Math.floor(zoom);
  if (v < 5) {
    return v - 1;
  }
  if (v < 7) {
    return 4;
  }
  if (v < 10) {
    return 5;
  }
  if (v < 13) {
    return 7;
  }
  return v;
}

function adjustLat(lat: number): number {
  if (lat < -90) {
    return -90; // Minimum latitude
  }
  if (lat > 90) {
    return 90; // Maximum latitude
  }
  return Math.round(lat * 1000000) / 1000000; // Round to 6 decimal places
}

function adjustLng(lng: number): number {
  if (lng < -180) {
    return -180; // Minimum longitude
  }
  if (lng > 180) {
    return 180; // Maximum longitude
  }
  return Math.round(lng * 1000000) / 1000000; // Round to 6 decimal places
}

function normalizeLng(lng: number): number {
  // Normalize longitude to the range [-180, 180]
  return ((lng + 180) % 360 + 360) % 360 - 180;
}

function App() {
  const { ecosystem_id } = useParams<{ ecosystem_id: string }>(); // Extract ecosystem_id from the URL
  const [supplyChainData, setSupplyChainData] = useState<any[]>([]); // State to store nodes
  const [edges, setEdges] = useState<any[]>([]); // State to store edges
  const mapRef = useRef<any>(null); // Reference to the Map instance

  const handleMapLoad = useCallback(() => {
    // if (mapRef.current) {
    //   const map = mapRef.current.getMap();
    //   console.log("Map has loaded:", map);

    //   map.on("styleimagemissing", (e) => {
    //     if (e.id === "arrow-icon") {
    //       console.log("Arrow icon is missing. Adding it dynamically...");
    //       const arrowIconUrl = "/arrow-icon.svg"; // URL of the arrow icon

    //       const load = async () => {
    //         try {
    //           console.log("Loading image from URL:", arrowIconUrl); // Log the URL being used
    //           const image = await map.loadImage(arrowIconUrl);
    //           if (!map.hasImage("arrow-icon")) {
    //             console.log("Adding arrow icon to map.", image);
    //             map.addImage("arrow-icon", image.data);
    //           }
    //         } catch (err) {
    //           console.error(`Failed to load image:`, err);
    //         }
    //       };

    //       load();
    //     }
    //   });
    // }
  }, []);

  const fetchSupplyChainData = async (boundingBox: {
    tl_lat: number;
    tl_lng: number;
    br_lat: number;
    br_lng: number;
    zoom: number;
  }) => {
    const { tl_lat, tl_lng, br_lat, br_lng, zoom } = boundingBox;
    try {
      const query = [
        `tl_lat=${adjustLat(tl_lat)}`,
        `tl_lng=${adjustLng(tl_lng)}`,
        `br_lat=${adjustLat(br_lat)}`,
        `br_lng=${adjustLng(br_lng)}`,
        `zoom=${adjustZommLevelV3(zoom)}`,
        `ecosystem_id=${ecosystem_id}`, // Use the ecosystem_id from the URL
      ].join("&");
      const response = await fetch(
        `http://localhost:8008/api/v3/supplychain?${query}`
      );
      const data = await response.json();
      console.log("Supply Chain Data:", data);
      setSupplyChainData(data.nodes || []); // Store nodes from the API response
      setEdges(data.edges || []); // Store edges from the API response
    } catch (error) {
      console.error("Error fetching supply chain data:", error);
    }
  };

  const handleZoomChange = useCallback(
    debounce(() => {
      if (!mapRef.current) return;

      const map = mapRef.current.getMap(); // Get the Map instance
      const bounds = map.getBounds(); // Get the bounding box
      const zoom = map.getZoom(); // Get the current zoom level

      const tl_lat = bounds.getNorth();
      const tl_lng = normalizeLng(bounds.getWest());
      const br_lat = bounds.getSouth();
      const br_lng = normalizeLng(bounds.getEast());

      console.log("Zoom changed:");
      console.log(`Zoom level: ${zoom}`);
      console.log(`Bounding Box: TL(${tl_lat}, ${tl_lng}), BR(${br_lat}, ${br_lng})`);

      // Fetch supply chain data using the bounding box and zoom level
      fetchSupplyChainData({
        tl_lat,
        tl_lng,
        br_lat,
        br_lng,
        zoom,
      });
    }, 200),
    []
  );

  const handleMapMove = useCallback(
    debounce(() => {
      if (!mapRef.current) return;

      const map = mapRef.current.getMap(); // Get the Map instance
      const bounds = map.getBounds(); // Get the bounding box
      const zoom = map.getZoom(); // Get the current zoom level

      // const tl_lat = bounds.getNorth();
      const tl_lat = 90;
      // const tl_lng = normalizeLng(bounds.getWest());
      const tl_lng = -180;
      // const br_lat = bounds.getSouth();
      const br_lat = -90;
      // const br_lng = normalizeLng(bounds.getEast());
      const br_lng = 180;

      console.log("Map moved:");
      console.log(`Zoom level: ${zoom}`);
      console.log(`Bounding Box: TL(${tl_lat}, ${tl_lng}), BR(${br_lat}, ${br_lng})`);

      // Fetch supply chain data using the bounding box and zoom level
      fetchSupplyChainData({
        tl_lat,
        tl_lng,
        br_lat,
        br_lng,
        zoom,
      });
    }, 200),
    []
  );

  // Create a hash map for quick node lookup
  const nodeMap = useMemo(() => {
    const map: Record<string, any> = {};
    supplyChainData.forEach((node) => {
      map[node.id] = node;
    });
    return map;
  }, [supplyChainData]);

  // Generate GeoJSON for edges
  const edgeGeoJSON = {
    type: "FeatureCollection",
    features: edges.map((edge) => {
      const sourceNode = nodeMap[edge.s];
      const targetNode = nodeMap[edge.t];
      if (sourceNode && targetNode) {
        return {
          type: "Feature",
          geometry: {
            type: "LineString",
            coordinates: [
              [sourceNode.lng, sourceNode.lat],
              [targetNode.lng, targetNode.lat],
            ],
          },
          properties: {
            source: edge.s,
            target: edge.t,
          },
        };
      }
      return null;
    }).filter(Boolean), // Remove null values
  };

  // Generate GeoJSON for arrowheads
  const arrowGeoJSON = {
    type: "FeatureCollection",
    features: edges.map((edge) => {
      const sourceNode = nodeMap[edge.s];
      const targetNode = nodeMap[edge.t];
      if (sourceNode && targetNode) {
        const dx = targetNode.lng - sourceNode.lng;
        const dy = targetNode.lat - sourceNode.lat;
        const angle = (Math.atan2(dy, dx) * 180) / Math.PI; // Calculate angle for rotation
        return {
          type: "Feature",
          geometry: {
            type: "Point",
            coordinates: [targetNode.lng, targetNode.lat],
          },
          properties: {
            rotation: angle, // Store rotation angle for the arrow
          },
        };
      }
      return null;
    }).filter(Boolean), // Remove null values
  };

  return (
    <Map
      ref={mapRef} // Attach the map reference
      initialViewState={{
        longitude: 106.6602, // Longitude for Ho Chi Minh City
        latitude: 10.7629,   // Latitude for Ho Chi Minh City
        zoom: 3,             // Adjusted zoom level for the city
      }}
      style={{ width: "100%", height: "100vh" }}
      // mapStyle="https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json"
      mapStyle="https://tiles.stadiamaps.com/styles/alidade_smooth.json"
      onLoad={handleMapLoad} // Triggered when the map has fully loaded
      onZoom={handleZoomChange}
      onMove={handleMapMove}
    >
      {/* Render circles based on supplyChainData */}
      {supplyChainData.map((node, index) => (
        <Marker
          key={`node-${index}`}
          longitude={node.lng}
          latitude={node.lat}
        >
          <div
            style={{
              position: "relative",
              width: 32,
              height: 32,
              backgroundColor: "blue",
              borderRadius: "50%",
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              color: "white",
              fontSize: "10px",
              fontWeight: "bold",
            }}
          >
            {node.value.size || 0} {/* Display size or a default value */}
          </div>
        </Marker>
      ))}

      {/* Render edges as lines */}
      <Source id="edges-source" type="geojson" data={edgeGeoJSON}>
        <Layer
          id="edges-layer"
          type="line"
          paint={{
            "line-color": "red",
            "line-width": 2,
            "line-opacity": 0.8,
          }}
          layout={{
            "line-cap": "round",
            "line-join": "round",
          }}
        />
      </Source>

      {/* Render arrowheads as symbols */}
      <Source id="arrows-source" type="geojson" data={arrowGeoJSON}>
        <Layer
          id="arrows-layer"
          type="symbol"
          layout={{
            "icon-image": "arrow-icon", // Use the loaded arrow icon
            "icon-size": 0.08, // Adjust the size of the arrow
            "icon-rotate": ["get", "rotation"], // Rotate the arrow based on the calculated angle
            "icon-anchor": "center",
          }}
        />
      </Source>
    </Map>
  );
}

export default function Root() {
  return (
    <Router>
      <Routes>
        <Route path="/:ecosystem_id" element={<App />} /> {/* Define the route with ecosystem_id */}
      </Routes>
    </Router>
  );
}
import React, { useCallback, useState, useRef, useEffect } from "react"; // eslint-disable-line
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import "./App.css";

import Map, { Marker } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";

const BASE_URL = "http://localhost:8000/api/track";
const MAP_STYLE = "https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json";
const MAP_CONF = {
  latitude: 55.7558,
  longitude: 37.6173,
  zoom: 10,
};

function debounce(func: (...args: any[]) => void, delay: number) {
  let timeout: NodeJS.Timeout; // eslint-disable-line
  return (...args: any[]) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), delay);
  };
}

function adjustLat(lat: number): number {
  if (lat < -90) {
    return -90;
  }
  if (lat > 90) {
    return 90;
  }
  return Math.round(lat * 1000000) / 1000000;
}

function adjustLng(lng: number): number {
  if (lng < -180) {
    return -180;
  }
  if (lng > 180) {
    return 180;
  }
  return Math.round(lng * 1000000) / 1000000;
}

function adjustZoom(zoom: number): number {
  return Math.min(Math.max(1, Math.floor(zoom * 0.6)), 15);
}

function getClusterSize(size?: number): number {
  if (!size) {
    return 15;
  }
  if (size > 4000) {
    return 250;
  }
  if (size > 3500) {
    return 225;
  }
  if (size > 3000) {
    return 200;
  }
  if (size > 2500) {
    return 18;
  }
  if (size > 2000) {
    return 160;
  }
  if (size > 1500) {
    return 140;
  }
  if (size > 1000) {
    return 130;
  }
  if (size > 800) {
    return 120;
  }
  if (size > 500) {
    return 110;
  }
  if (size > 250) {
    return 100;
  }
  if (size > 125) {
    return 90;
  }
  if (size > 100) {
    return 80;
  }
  if (size > 75) {
    return 70;
  }
  if (size > 1) {
    return 60;
  }
  return 15;
}

function App() {
  const [locations, setLocations] = useState<any[]>([]);
  const mapRef = useRef<any>(null);

  const fetchLocations = async () => {
    if (!mapRef.current) {
      return;
    }

    const map = mapRef.current.getMap();
    const bounds = map.getBounds();
    const zoom = adjustZoom(map.getZoom());
    const latMin = bounds.getSouth();
    const latMax = bounds.getNorth();
    const lngMin = bounds.getWest();
    const lngMax = bounds.getEast();

    console.log(`Zoom: ${zoom}`);

    try {
      const query = [
        `latMin=${adjustLat(latMin)}`,
        `latMax=${adjustLat(latMax)}`,
        `lngMin=${adjustLng(lngMin)}`,
        `lngMax=${adjustLng(lngMax)}`,
        `zoom=${zoom}`,
      ].join("&");

      const response = await fetch(`${BASE_URL}/locations?${query}`);
      const data = await response.json();
      console.log("Locations:", data.items);
      setLocations(data.items || []);
    } catch (error) {
      console.error("Error fetching locations:", error);
    }
  };

  const handleZoomChange = useCallback(debounce(fetchLocations, 200), []);
  const handleMapMove = useCallback(debounce(fetchLocations, 200), []);
  
  useEffect(() => {
    const intervalId = setInterval(fetchLocations, 1000);
    return () => clearInterval(intervalId); // Cleanup on unmount
  }, []);

  return (
    <Map
      ref={mapRef}
      initialViewState={MAP_CONF}
      style={{ width: "100%", height: "100vh" }}
      mapStyle={MAP_STYLE}
      onZoom={handleZoomChange}
      onMove={handleMapMove}
    >
      {locations.map((node) => (
        <Marker
          key={node.id}
          longitude={node.loc[1]}
          latitude={node.loc[0]}
        >
          <div
            style={{
              position: "relative",
              width: getClusterSize(node.value),
              height: getClusterSize(node.value),
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
            {node.value && node.value > 1 ? node.value : ''} {/* Display size or a default value */}
          </div>
        </Marker>
      ))}
    </Map>
  );
}

export default function Root() {
  return (
    <Router>
      <Routes>
        <Route path="/locations" element={<App />} />
      </Routes>
    </Router>
  );
}

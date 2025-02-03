<template>
  <div class="w-full h-full bg-white shadow rounded-lg p-4">
    <!-- <h2 class="text-xl font-semibold mb-4">Rioters Locations</h2> -->
    <div ref="mapContainer" class="w-full h-[100vh] rounded-lg" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, defineProps, onBeforeUnmount } from "vue";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { getImageUrl } from "../utils/imageHandling";
import Supercluster from "supercluster";

const MAPBOX_ACCESS_TOKEN = process.env.VUE_APP_MAPBOX_ACCESS_TOKEN;
const createPopupContent = (rioter) => {
  return `
    <div class="p-2">
      <div class="flex items-center mb-2">
        <img 
          src="${getImageUrl(rioter.photo_name)}"
          alt="${rioter.first_name} ${rioter.last_name}"
          class="h-12 w-12 rounded-full object-cover mr-2"
          onerror="this.src='${getImageUrl()}'"
        />
        <div>
          <strong>${rioter.first_name} ${rioter.last_name}</strong><br>
          ${rioter.city ? rioter.city + ", " : ""}${rioter.state || ""}
        </div>
      </div>
      ${rioter.charges ? `<small class="text-gray-600">${rioter.charges}</small>` : ""}
    </div>
  `;
};
const handleResize = () => {
  if (map) {
    map.resize(); // Ensures the map adjusts properly
  }
};

onMounted(() => {
  initializeMap();
  window.addEventListener("resize", handleResize); // Listen for resize
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize); // Cleanup
});
const props = defineProps({
  rioters: {
    type: Array,
    required: true,
  },
  bounds: {
    type: Array,
    default: null,
  },
});

let map = null;
const mapContainer = ref(null);
const markers = ref([]);

const initializeMap = () => {
  if (!map) {
    mapboxgl.accessToken = MAPBOX_ACCESS_TOKEN;
    map = new mapboxgl.Map({
      container: mapContainer.value,
      style: "mapbox://styles/mapbox/streets-v11",
      ...(props.bounds
        ? {
            bounds: props.bounds,
            fitBoundsOptions: { padding: 50 },
          }
        : {
            center: [-98.5795, 39.8283],
            zoom: 4,
          }),
    });

    map.addControl(new mapboxgl.NavigationControl());
    updateMarkers();
  }
};

// const clearMarkers = () => {
//   markers.value.forEach((marker) => marker.remove());
//   markers.value = [];
// };

const updateMarkers = () => {
  if (!map) return;

  // Clear existing markers
  markers.value.forEach((marker) => marker.remove());
  markers.value = [];

  // Create GeoJSON FeatureCollection from rioters data
  const geojsonData = {
    type: "FeatureCollection",
    features: props.rioters
      .filter((r) => r.latitude && r.longitude)
      .map((rioter) => ({
        type: "Feature",
        properties: {
          id: rioter.id,
          first_name: rioter.first_name,
          last_name: rioter.last_name,
          city: rioter.city,
          state: rioter.state,
          charges: rioter.charges,
          photo_name: rioter.photo_name,
        },
        geometry: {
          type: "Point",
          coordinates: [parseFloat(rioter.longitude), parseFloat(rioter.latitude)],
        },
      })),
  };

  // Check for overlapping markers by clustering logic
  const cluster = new Supercluster({
    radius: 50,
    maxZoom: 14,
  }).load(geojsonData.features);

  // Add markers or clusters
  geojsonData.features.forEach((feature) => {
    const [lng, lat] = feature.geometry.coordinates;

    const clusterPoints = cluster.getClusters([lng, lat, lng, lat], map.getZoom());

    if (clusterPoints.length > 1) {
      // Handle overlapping markers by creating a cluster marker
      const clusterMarker = new mapboxgl.Marker({ color: "red" })
        .setLngLat([lng, lat])
        .setPopup(
          new mapboxgl.Popup().setHTML(
            `<strong>${clusterPoints.length} rioters here</strong>`
          )
        )
        .addTo(map);

      markers.value.push(clusterMarker);
    } else {
      // Normal marker for individual rioters
      const marker = new mapboxgl.Marker()
        .setLngLat([lng, lat])
        .setPopup(new mapboxgl.Popup().setHTML(createPopupContent(feature.properties)))
        .addTo(map);

      markers.value.push(marker);
    }
  });

  // Auto-zoom to markers if bounds exist
  if (props.bounds && map) {
    map.fitBounds(props.bounds, {
      padding: 50,
      maxZoom: 12,
      duration: 1000,
    });
  }
};

watch(() => props.rioters, updateMarkers, { deep: true });

watch(
  () => props.bounds,
  (newBounds) => {
    if (map && newBounds) {
      map.fitBounds(newBounds, {
        padding: 50,
        maxZoom: 12,
        duration: 1000,
      });
    } else if (map) {
      map.flyTo({
        center: [-98.5795, 39.8283],
        zoom: 3,
      });
    }
  },
  { immediate: true }
);
// onMounted(initializeMap);
onBeforeUnmount(() => {
  if (map) map.remove();
});
</script>

<style>
.map-container {
  width: w-full;
  height: h-full;
  /* min-height: 400px; Prevents the map from disappearing in small containers */
  position: relative;
}
@media (max-width: 1024px) {
  .lg\:rounded-l-lg {
    border-radius: 0;
  }
  .lg\:static {
    position: static;
  }
}
</style>

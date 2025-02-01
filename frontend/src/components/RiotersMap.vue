<template>
  <div class="w-full h-full bg-white shadow rounded-lg p-4">
    <!-- <h2 class="text-xl font-semibold mb-4">Rioters Locations</h2> -->
    <div
      ref="mapContainer"
      class="w-full h-[100vh] rounded-lg"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, defineProps, onBeforeUnmount } from "vue";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { getImageUrl } from "../utils/imageHandling";

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

const clearMarkers = () => {
  markers.value.forEach((marker) => marker.remove());
  markers.value = [];
};

const updateMarkers = () => {
  clearMarkers();
  // Check if we have valid coordinates
  const validRioters = props.rioters.filter(
    (r) => r.latitude && r.longitude && !isNaN(r.latitude) && !isNaN(r.longitude)
  );

  if (validRioters.length === 0 && map) {
    map.flyTo({
      center: [-98.5795, 39.8283],
      zoom: 3,
    });
    return;
  }
  props.rioters.forEach((rioter) => {
    if (rioter.latitude && rioter.longitude) {
      const lat = parseFloat(rioter.latitude);
      const lng = parseFloat(rioter.longitude);
      if (isNaN(lat) || isNaN(lng)) {
        console.error("Invalid coordinates for rioter:", rioter.id);
        return;
      }

      const marker = new mapboxgl.Marker()
        .setLngLat([rioter.longitude, rioter.latitude])
        .setPopup(
          new mapboxgl.Popup().setHTML(createPopupContent(rioter)) // Here's the fix
        )
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

<template>
  <div class="w-full bg-white shadow rounded-lg p-4">
    <h2 class="text-xl font-semibold mb-4">Rioters Locations</h2>
    <div ref="mapContainer" class="h-96 rounded-lg"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, defineProps, onBeforeUnmount } from "vue";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { getImageUrl } from '../utils/imageHandling';

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
          ${rioter.city ? rioter.city + ', ' : ''}${rioter.state || ''}
        </div>
      </div>
      ${rioter.charges ? `<small class="text-gray-600">${rioter.charges}</small>` : ''}
    </div>
  `;
};

const props = defineProps({
  rioters: {
    type: Array,
    required: true
  },
  bounds: {
    type: Array,
    default: null
  }
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
      ...(props.bounds ? {
        bounds: props.bounds,
        fitBoundsOptions: { padding: 50 }
      } : {
        center: [-98.5795, 39.8283],
        zoom: 4
      })
    });

    map.addControl(new mapboxgl.NavigationControl());
    updateMarkers();
  }
};

const clearMarkers = () => {
  markers.value.forEach(marker => marker.remove());
  markers.value = [];
};

const updateMarkers = () => {
  clearMarkers();
  props.rioters.forEach((rioter) => {
    if (rioter.latitude && rioter.longitude) {
      const marker = new mapboxgl.Marker()
        .setLngLat([rioter.longitude, rioter.latitude])
        .setPopup(
          new mapboxgl.Popup().setHTML(`
            <div class="text-sm">
              <strong>${rioter.first_name} ${rioter.last_name}</strong><br>
              ${rioter.city}, ${rioter.state}
            </div>
          `)
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
      duration: 1000
    });
  }
};

watch(() => props.rioters, updateMarkers, { deep: true });

watch(() => props.bounds, (newBounds) => {
  if (map && newBounds) {
    map.fitBounds(newBounds, {
      padding: 50,
      maxZoom: 12,
      duration: 1000
    });
  }
});

onMounted(initializeMap);
onBeforeUnmount(() => {
  if (map) map.remove();
});
</script>

<style>
.map-container {
  width: 100%;
  height: 400px;
  border-radius: 0.5rem;
  overflow: hidden;
}
</style>
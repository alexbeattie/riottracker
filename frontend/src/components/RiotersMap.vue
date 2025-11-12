<template>
  <div class="bg-white shadow rounded-lg">
    <!-- <h2 class="text-xl font-semibold mb-4">Rioters Locations</h2> -->
    <div ref="mapContainer" class="w-full h-[100vh] rounded-lg" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, defineProps, onBeforeUnmount, defineEmits } from "vue";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { getImageUrl } from "../utils/imageHandling";
import { nextTick } from "vue";
const emit = defineEmits(["marker-click", "center-map"]);
const MAPBOX_ACCESS_TOKEN = process.env.VUE_APP_MAPBOX_ACCESS_TOKEN;
const createPopupContent = (rioter) => {
  return `
    <div class="p-2">
      <div class="flex items-center">
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
    
      
    
      </div>
  `;
};
const handleResize = () => {
  if (map) {
    map.resize(); // Ensures the map adjusts properly
  }
};

onMounted(() => {
  // Suppress unhandled promise rejections from browser extensions/Mapbox workers
  window.addEventListener('unhandledrejection', (event) => {
    const errorMsg = event.reason?.message || event.reason?.toString() || '';
    if (errorMsg.includes('Could not establish connection') || 
        errorMsg.includes('Receiving end does not exist')) {
      event.preventDefault(); // Prevent the error from showing in console
      return;
    }
  });
  
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
  selectedRioter: {
    type: Object,
    default: null,
  },
});

let map = null;
const mapContainer = ref(null);
const markers = ref([]);

const initializeMap = () => {
  if (!map) {
    if (!MAPBOX_ACCESS_TOKEN) {
      console.error("Mapbox access token is not set. Please set VUE_APP_MAPBOX_ACCESS_TOKEN in your .env file");
      return;
    }
    
    mapboxgl.accessToken = MAPBOX_ACCESS_TOKEN;
    
    try {
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

      // Handle map errors gracefully
      map.on('error', (e) => {
        // Suppress connection errors from browser extensions
        if (e.error?.message?.includes('Could not establish connection')) {
          return;
        }
        console.error('Map error:', e.error);
      });

      map.addControl(new mapboxgl.NavigationControl());
      
      // Wait for map to load before updating markers
      map.on('load', () => {
        updateMarkers();
      });
      
      // Also update markers immediately in case map is already loaded
      updateMarkers();
    } catch (err) {
      console.error("Error initializing map:", err);
    }
  }
};
const flyToMarker = (rioter) => {
  if (!map || !rioter || !rioter.latitude || !rioter.longitude) return;

  console.log("🛫 Flying to:", rioter.first_name, rioter.last_name);
  map.flyTo({
    center: [rioter.longitude, rioter.latitude],
    zoom: 10,
    essential: true,
  });
};

// const clearMarkers = () => {
//   markers.value.forEach((marker) => marker.remove());
//   markers.value = [];
// };

const updateMarkers = () => {
  if (!map) return;

  console.log(`🗺️ Updating markers for ${props.rioters.length} rioters`);

  // Clear existing markers
  markers.value.forEach((marker) => marker.remove());
  markers.value = [];

  // Filter rioters with valid coordinates
  const validRioters = props.rioters.filter(r => {
    const hasCoords = r.latitude && r.longitude && 
                      !isNaN(parseFloat(r.latitude)) && 
                      !isNaN(parseFloat(r.longitude));
    if (!hasCoords) {
      console.warn(`⚠️ Rioter ${r.id} (${r.first_name} ${r.last_name}) missing coordinates`);
    }
    return hasCoords;
  });

  console.log(`✅ ${validRioters.length} rioters have valid coordinates`);

  // Group rioters by coordinates
  const groupedRioters = validRioters.reduce((acc, rioter) => {
    const lat = parseFloat(rioter.latitude);
    const lng = parseFloat(rioter.longitude);
    const key = `${lat},${lng}`;
    if (!acc[key]) {
      acc[key] = {
        rioters: [],
        coordinates: [lng, lat],
        city: rioter.city,
        state: rioter.state,
      };
    }
    acc[key].rioters.push(rioter);
    return acc;
  }, {});

  const seenCoords = new Map();

  // Call it inside the watcher
  watch(
    () => props.selectedRioter,
    async (newSelectedRioter) => {
      if (!newSelectedRioter) {
        return; // No rioter selected, nothing to do
      }
      
      if (markers.value.length === 0) {
        // Markers not loaded yet, wait a bit and try again
        await new Promise(resolve => setTimeout(resolve, 500));
        if (markers.value.length === 0) {
          console.warn("⚠️ Markers not loaded yet, cannot center on rioter");
          return;
        }
      }

      console.log("📍 Centering map on:", newSelectedRioter);

      // Ensure the selected rioter has valid coordinates
      if (!newSelectedRioter.latitude || !newSelectedRioter.longitude) {
        console.warn("⚠️ No valid lat/lng for:", newSelectedRioter);
        return;
      }

      // 🔹 Fly to marker
      if (map) {
        console.log(
          "🛫 Flying to:",
          newSelectedRioter.first_name,
          newSelectedRioter.last_name
        );
        map.flyTo({
          center: [newSelectedRioter.longitude, newSelectedRioter.latitude],
          zoom: 10, // Adjust zoom level if necessary
          essential: true,
        });
      }

      await nextTick(); // ✅ Wait for DOM updates

      // 🔹 Close all existing popups
      markers.value.forEach((marker) => marker.getPopup().remove());

      // 🔹 Find the correct marker
      const marker = markers.value.find((m) => m._rioter?.id === newSelectedRioter.id);
      if (marker) {
        console.log(
          "✅ Found marker for:",
          newSelectedRioter.first_name,
          newSelectedRioter.last_name
        );

        // Ensure the popup is set properly
        const popup = new mapboxgl.Popup({ maxWidth: "300px" })
          .setLngLat([newSelectedRioter.longitude, newSelectedRioter.latitude])
          .setHTML(createPopupContent(newSelectedRioter))
          .addTo(map);

        marker.setPopup(popup);
        marker.togglePopup(); // ✅ Open the popup
      } else {
        console.warn("⚠️ Marker not found for selected rioter:", newSelectedRioter);
      }
    }
  );

  Object.values(groupedRioters).forEach((group) => {
    const isCluster = group.rioters.length > 1;
    const baseCoordinates = group.coordinates;

    group.rioters.forEach((rioter) => {
      let [lng, lat] = baseCoordinates;

      // Apply a slight jitter if this coordinate is already seen
      const key = `${lng},${lat}`;
      if (seenCoords.has(key)) {
        const jitterAmount = 0.0006 + Math.random() * 0.008; // Adjust jitter dynamically
        const angle = (Math.random() * 360 * Math.PI) / 180; // Random angle
        lng += Math.cos(angle) * jitterAmount;
        lat += Math.sin(angle) * jitterAmount;
      }

      seenCoords.set(`${lng},${lat}`, true);

      // Create marker
      const marker = new mapboxgl.Marker({
        color: isCluster ? "#f00" : "#4a4a4a",
        scale: isCluster ? 1.2 : 1,
      })
        .setLngLat([lng, lat])
        .setPopup(
          new mapboxgl.Popup({
            maxWidth: "300px",
          }).setHTML(createPopupContent(rioter))
        )
        .addTo(map);

      marker._rioter = rioter;

      marker.getElement().addEventListener("click", () => {
        emit("marker-click", rioter);
      });

      markers.value.push(marker);
    });
  });

  console.log(`📍 Created ${markers.value.length} markers on map`);

  // Fit bounds if provided
  if (props.bounds && map) {
    map.fitBounds(props.bounds, {
      padding: 50,
      maxZoom: 12,
      duration: 1000,
    });
  }
};
// Fit bounds if provided

watch(() => props.rioters, updateMarkers, { deep: true });

// watch(
//   () => props.selectedRioter,
//   async (newSelectedRioter) => {
//     if (!newSelectedRioter || markers.value.length === 0) return;

//     console.log(
//       "Trying to open popup for:",
//       newSelectedRioter.first_name,
//       newSelectedRioter.last_name
//     );

//     await nextTick(); // ✅ Wait for DOM updates

//     // Close all popups first
//     markers.value.forEach((marker) => marker.getPopup().remove());

//     // Find the correct marker
//     const marker = markers.value.find((m) => m._rioter?.id === newSelectedRioter.id);
//     if (marker) {
//       console.log(
//         "✅ Found marker for:",
//         newSelectedRioter.first_name,
//         newSelectedRioter.last_name
//       );

//       // Ensure the popup is re-created and opened properly
//       const popup = new mapboxgl.Popup({ maxWidth: "300px" })
//         .setLngLat([newSelectedRioter.longitude, newSelectedRioter.latitude])
//         .setHTML(createPopupContent(newSelectedRioter))
//         .addTo(map);

//       marker.setPopup(popup); // ✅ Ensure popup is set on the marker
//       marker.togglePopup(); // ✅ Open it
//     } else {
//       console.warn("⚠️ Marker not found for selected rioter:", newSelectedRioter);
//     }
//   }
// );

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
        zoom: 4,
      });
    }
  },
  { immediate: true }
);
const getMapCenter = () => {
  if (!map) {
    console.warn("Map not initialized yet");
    return null;
  }
  try {
    // Check if map is loaded
    if (!map.loaded()) {
      console.warn("Map not loaded yet");
      // Return current center even if not fully loaded
    }
    const center = map.getCenter();
    if (!center) {
      console.warn("Map center is null");
      return null;
    }
    
    const lat = typeof center.lat === 'function' ? center.lat() : center.lat;
    const lng = typeof center.lng === 'function' ? center.lng() : center.lng;
    
    if (typeof lat !== 'number' || typeof lng !== 'number' || isNaN(lat) || isNaN(lng)) {
      console.warn("Map center coordinates invalid:", { lat, lng });
      return null;
    }
    
    console.log("Map center retrieved:", { lat, lng });
    return {
      lat: lat,
      lng: lng
    };
  } catch (err) {
    console.error("Error getting map center:", err);
    return null;
  }
};

defineExpose({ flyToMarker, getMapCenter });

// onMounted(initializeMap);
onBeforeUnmount(() => {
  if (map) map.remove();
});
</script>

<style>
.map-container {
  width: w-full;
  height: h-full;
  position: relative;
}

.cluster-marker {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  z-index: 1;
}

.mapboxgl-popup-content {
  padding: 0 !important;
  border-radius: 8px !important;
}

@media (max-width: 1024px) {
  .lg\:rounded-l-lg {
    border-radius: 0;
  }
  .lg\:static {
    position: static;
  }
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

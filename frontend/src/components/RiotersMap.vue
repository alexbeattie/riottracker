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

  // Group rioters by coordinates
  const groupedRioters = props.rioters.reduce((acc, rioter) => {
    if (!rioter.latitude || !rioter.longitude) return acc;

    const key = `${rioter.latitude},${rioter.longitude}`;
    if (!acc[key]) {
      acc[key] = {
        rioters: [],
        coordinates: [parseFloat(rioter.longitude), parseFloat(rioter.latitude)],
        city: rioter.city,
        state: rioter.state,
      };
    }
    acc[key].rioters.push(rioter);
    return acc;
  }, {});

  // Create markers for each group
  Object.values(groupedRioters).forEach((group) => {
    const isCluster = group.rioters.length > 1;

    // Create custom popup content for clusters
    const createClusterPopup = (rioters) => {
      return `
        <div class="p-4">
          <h3 class="font-bold mb-2">${group.city}, ${group.state}</h3>
          <p class="mb-2">${rioters.length} rioters at this location</p>
          <div class="max-h-60 overflow-y-auto">
            ${rioters
              .map(
                (rioter) => `
              <div class="flex items-center mb-2 border-b pb-2">
                <img 
                  src="${getImageUrl(rioter.photo_name)}"
                  alt="${rioter.first_name} ${rioter.last_name}"
                  class="h-8 w-8 rounded-full object-cover mr-2"
                  onerror="this.src='${getImageUrl()}'"/>
                <div>
                  <div class="font-semibold">${rioter.first_name} ${
                  rioter.last_name
                }</div>
                  ${
                    rioter.charges
                      ? `<small class="text-gray-600">${rioter.charges}</small>`
                      : ""
                  }
                </div>
              </div>
            `
              )
              .join("")}
          </div>
        </div>
      `;
    };

    // Create marker with appropriate style and popup
    const marker = new mapboxgl.Marker({
      color: isCluster ? "#f00" : "#4a4a4a",
      scale: isCluster ? 1.2 : 1,
    })
      .setLngLat(group.coordinates)
      .setPopup(
        new mapboxgl.Popup({
          maxWidth: "300px",
        }).setHTML(
          isCluster
            ? createClusterPopup(group.rioters)
            : createPopupContent(group.rioters[0])
        )
      )
      .addTo(map);

    // Add cluster count if needed
    if (isCluster) {
      const el = document.createElement("div");
      el.className = "cluster-marker";
      el.style.backgroundColor = "white";
      el.style.borderRadius = "50%";
      el.style.width = "20px";
      el.style.height = "20px";
      el.style.display = "flex";
      el.style.alignItems = "center";
      el.style.justifyContent = "center";
      el.style.position = "absolute";
      el.style.top = "-10px";
      el.style.right = "-10px";
      el.style.border = "2px solid #f00";
      el.style.fontSize = "12px";
      el.style.fontWeight = "bold";
      el.innerText = group.rioters.length;

      marker.getElement().appendChild(el);
    }

    markers.value.push(marker);
  });

  // Fit bounds if provided
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

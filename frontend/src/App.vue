<!-- src/App.vue -->
<template>
  <div class="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
    <div class="max-w-7xl mx-auto">
      <h1 class="text-3xl font-bold text-gray-900 text-center mb-8">
        Rioters Database
      </h1>
      <div class="mb-4 text-sm text-gray-600">
        <span class="font-semibold">{{ filteredRioters.length }}</span>
        {{ filteredRioters.length === 1 ? "result" : "results" }} found
        <span v-if="currentFilters.state" class="ml-2">
          in <span class="font-medium">{{ currentFilters.state }}</span>
        </span>
        <span v-if="fetchMode === 'nearby'" class="ml-2">
          within 50 km of
          <span v-if="userLocation">your location</span>
          <span v-else>Los Angeles</span>
        </span>
      </div>
      <search-filters @filters-changed="handleFiltersChange" />

      <div v-if="loading" class="flex justify-center items-center py-12">
        <div
          class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"
        >
          <!-- Add the map component -->
        </div>
      </div>

      <div v-if="error" class="bg-red-50 border-l-4 border-red-400 p-4 mb-8">
        <div class="flex">
          <div class="ml-3">
            <p class="text-sm text-red-700">{{ error }}</p>
          </div>
        </div>
      </div>
      <div class="mt-8 relative">
        <div
          v-if="loading"
          class="absolute inset-0 bg-gray-100/50 z-10 flex items-center justify-center"
        >
          <div
            class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"
          ></div>
        </div>
        <rioters-map
          :rioters="filteredRioters"
          :bounds="mapBounds"
          :key="mapKey"
        />
      </div>
      <button
        @click="toggleFetchMode"
        :disabled="loading"
        class="px-4 py-2 rounded transition-all relative"
        :class="{
          'bg-blue-500 text-white hover:bg-blue-600': fetchMode === 'all',
          'bg-green-500 text-white hover:bg-green-600': fetchMode === 'nearby',
          'opacity-50 cursor-not-allowed': loading,
        }"
      >
        <span v-if="!loading">
          {{ fetchMode === "all" ? "Show Nearby Rioters" : "Show All Rioters" }}
        </span>
        <span v-else class="flex items-center">
          <span class="animate-spin mr-2">⟳</span>
          Loading...
        </span>
      </button>
      <ul
        v-if="filteredRioters.length > 0"
        class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6"
      >
        <li
          v-for="rioter in filteredRioters"
          :key="rioter.id"
          class="bg-white shadow rounded-lg overflow-hidden hover:shadow-lg transition-shadow duration-300"
        >
          <div class="p-6">
            <div class="flex items-center space-x-4 mb-4">
              <img
    :src="getImageUrl(rioter.photo_name)"
    @error="handleImageError"
    class="h-24 w-24 rounded-full object-cover border-2 border-gray-200"
  />
              <div>
                <h2 class="text-xl font-semibold text-gray-900">
                  {{ rioter.first_name }}
                  {{ rioter.middle_name ? rioter.middle_name + " " : ""
                  }}{{ rioter.last_name }}
                </h2>
                <p v-if="rioter.age" class="text-sm text-gray-600">
                  Age: {{ rioter.age }}
                </p>
              </div>
            </div>

            <div class="space-y-3">
              <div v-if="rioter.city || rioter.state" class="text-sm">
                <span class="font-medium">Location:</span>
                {{ [rioter.city, rioter.state].filter(Boolean).join(", ") }}
              </div>

              <div v-if="rioter.summary" class="text-sm">
                <span class="font-medium">Summary:</span>
                <p class="mt-1 text-gray-600">{{ rioter.summary }}</p>
              </div>

              <div v-if="rioter.jurisdiction" class="text-sm">
                <span class="font-medium">Jurisdiction:</span>
                <span class="text-gray-600">{{ rioter.jurisdiction }}</span>
              </div>

              <div v-if="rioter.charges" class="text-sm">
                <span class="font-medium">Charges:</span>
                <p class="mt-1 text-gray-600">{{ rioter.charges }}</p>
              </div>

              <div v-if="rioter.case_status" class="text-sm">
                <span class="font-medium">Case Status:</span>
                <p class="mt-1 text-gray-600">{{ rioter.case_status }}</p>
              </div>

              <div v-if="rioter.case_updates" class="text-sm">
                <span class="font-medium">Case Updates:</span>
                <p class="mt-1 text-gray-600">{{ rioter.case_updates }}</p>
              </div>

              <!-- Tags Section -->
              <div class="flex flex-wrap gap-2 mt-4">
                <span
                  v-if="rioter.violence_assault"
                  class="px-2 py-1 bg-red-100 text-red-800 text-xs rounded-full"
                >
                  Violence/Assault
                </span>
                <span
                  v-if="rioter.conspiracy"
                  class="px-2 py-1 bg-orange-100 text-orange-800 text-xs rounded-full"
                >
                  Conspiracy
                </span>
                <span
                  v-if="rioter.property"
                  class="px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded-full"
                >
                  Property Damage
                </span>
                <span
                  v-if="rioter.military_le"
                  class="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full"
                >
                  Military/LE
                </span>
                <span
                  v-if="rioter.extremist"
                  class="px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full"
                >
                  Extremist
                </span>
                <span
                  v-if="rioter.sentenced"
                  class="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full"
                >
                  Sentenced
                </span>
              </div>

              <div class="mt-4">
                <a
                  v-if="rioter.charges_link"
                  :href="rioter.charges_link"
                  target="_blank"
                  class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-blue-700 bg-blue-50 hover:bg-blue-100"
                >
                  View Charges Source
                  <svg
                    class="ml-2 h-4 w-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                    />
                  </svg>
                </a>
              </div>
            </div>
          </div>
        </li>
      </ul>

      <div
        v-else-if="!loading"
        class="bg-white shadow rounded-lg p-6 text-center"
      >
        <p class="text-gray-500">No results found matching your filters.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import SearchFilters from "./components/SearchFilters.vue";
import RiotersMap from "./components/RiotersMap.vue";
// import RioterImage from "./components/RioterImage.vue";
// import { getImageUrl } from "./utils/imageHandling";
import api from "./api";
// Function to get image URL
// Add this to your script setup
// And update the handleImageError function:
const utils = {
  getImageUrl: (photoName) => {
    const baseUrl = "http://localhost:8080";
    if (!photoName?.trim()) {
      return `${baseUrl}/photos/placeholder.jpg`;
    }
    return `${baseUrl}/photos/${encodeURIComponent(photoName)}`;
  },

  handleImageError: (event) => {
    const placeholder = `http://localhost:8080/photos/placeholder.jpg`;
    if (event.target.src !== placeholder) {
      event.target.src = placeholder;
    }
  }
};

const mapBounds = computed(() => {
  if (filteredRioters.value.length === 0) return null;

  const validRioters = filteredRioters.value.filter(
    (r) =>
      r.latitude && r.longitude && !isNaN(r.latitude) && !isNaN(r.longitude)
  );

  if (validRioters.length === 0) return null;

  const lngs = validRioters.map((r) => r.longitude);
  const lats = validRioters.map((r) => r.latitude);

  return [
    [Math.min(...lngs), Math.min(...lats)], // SW
    [Math.max(...lngs), Math.max(...lats)], // NE
  ];
});

const toggleFetchMode = () => {
  // Reset filters when changing modes
  currentFilters.value = {
    searchText: "",
    state: "",
    charges: "",
    status: "",
    isNewOrUpdated: false,
    affiliations: {
      military_le: false,
      extremist: false,
      sentenced: false,
      commuted: false,
    },
  };

  fetchMode.value = fetchMode.value === "all" ? "nearby" : "all";
  fetchRioters();
};

// const getImageUrl = (photoName) => {
//   const baseUrl = process.env.VUE_APP_API_BASE_URL || "http://localhost:8080";

//   if (!photoName?.trim()) {
//     return `${baseUrl}/photos/placeholder.jpg`;
//   }

//   return `${baseUrl}/photos/${encodeURIComponent(photoName)}`;
// };
// Handle missing images
// const handleImageError = (event) => {
//   const placeholder = `${process.env.VUE_APP_API_BASE_URL}/photos/placeholder.jpg`;
//   if (event.target.src !== placeholder) {
//     event.target.src = placeholder;
//   }
// };
// const getImageUrl = (photoName) => {
//   // Always use placeholder if no photo name or by default
//   if (!photoName) return "/photos/placeholder.jpg";

//   // Even if we have a photo name, start with placeholder as default
//   let imageUrl = "/photos/placeholder.jpg";

//   // Only if we have a matching photo in the photos directory, use that instead
//   if (photoName) {
//     imageUrl = `/photos/${photoName}`;
//   }

//   return imageUrl;
// };
const getImageUrl = utils.getImageUrl;
const handleImageError = utils.handleImageError;

const rioters = ref([]);
const loading = ref(true);
const error = ref(null);
const currentFilters = ref({
  searchText: "",
  state: "",
  charges: "",
  status: "",
  isNewOrUpdated: false,
  affiliations: {
    military_le: false,
    extremist: false,
    sentenced: false,
    commuted: false,
  },
});

const filteredRioters = computed(() => {
  return rioters.value.filter((rioter) => {
    // Text search
    if (currentFilters.value.searchText) {
      const searchText = currentFilters.value.searchText.toLowerCase();
      const searchMatch =
        `${rioter.first_name} ${rioter.last_name}`
          .toLowerCase()
          .includes(searchText) ||
        rioter.summary?.toLowerCase().includes(searchText) ||
        false ||
        rioter.charges?.toLowerCase().includes(searchText) ||
        false;

      if (!searchMatch) return false;
    }

    // State filter
    if (currentFilters.value.state && rioter.state) {
      if (
        rioter.state.toLowerCase() !== currentFilters.value.state.toLowerCase()
      ) {
        return false;
      }
    }

    // Charges filter
    if (currentFilters.value.charges) {
      const chargeType = currentFilters.value.charges;
      switch (chargeType) {
        case "violence_assault":
          if (!rioter.violence_assault) return false;
          break;
        case "conspiracy":
          if (!rioter.conspiracy) return false;
          break;
        case "property":
          if (!rioter.property) return false;
          break;
      }
    }

    // Status filter
    if (currentFilters.value.status && rioter.case_status) {
      const status = currentFilters.value.status.toLowerCase();
      if (!rioter.case_status.toLowerCase().includes(status)) {
        return false;
      }
    }

    // Affiliations filters
    const activeAffiliations = Object.entries(currentFilters.value.affiliations)
      .filter(([, value]) => value)
      .map(([key]) => key);

    if (activeAffiliations.length > 0) {
      for (const affiliation of activeAffiliations) {
        switch (affiliation) {
          case "military_le":
            if (!rioter.military_le) return false;
            break;
          case "extremist":
            if (!rioter.extremist) return false;
            break;
          case "sentenced":
            if (!rioter.sentenced) return false;
            break;
          case "commuted":
            if (!rioter.commuted) return false;
            break;
        }
      }
    }

    return true;
  });
});
const handleFiltersChange = (filters) => {
  currentFilters.value = filters;

  // Auto-zoom to state if state filter applied
  if (filters.state) {
    const stateCenters = {
      ca: [-119.417931, 37.184092],
      tx: [-99.359349, 31.816038],
      ny: [-75.144424, 43.156168],
      // Add more states as needed
    };

    const center = stateCenters[filters.state.toLowerCase()];
    if (center) {
      mapBounds.value = [
        [center[0] - 2, center[1] - 1], // SW
        [center[0] + 2, center[1] + 1], // NE
      ];
    }
  }
};

const fetchMode = ref("all"); // Default to fetching all
onMounted(() => {
  fetchRioters();
});
const userLocation = ref(null);
const mapKey = ref(0);

const fetchRioters = async () => {
  loading.value = true;
  error.value = null;

  try {
    let response;
    userLocation.value = null; // Reset location state

    if (fetchMode.value === "all") {
      response = await api.get("/rioters");
    } else {
      let coords;
      try {
        coords = await getCoordinates();
        userLocation.value = coords; // Store successful location
      } catch (err) {
        console.error("Error getting coordinates:", err);
        userLocation.value = { lat: 34.052235, lng: -118.243683 };
      }

      const defaultRadius = 50000; // 50 km radius

      response = await api.get("/rioters/nearby", {
        params: {
          lng: coords.lng, // Note parameter order
          lat: coords.lat,
          radius: defaultRadius,
        },
      });
    }

    // 🔹 Debug API response
    console.log("API Full Response:", response);
    console.log("Response Data:", response.data);

    // Check if response.data is missing or not an array
    if (!response.data || typeof response.data !== "object") {
      throw new Error(
        "Unexpected response format: " + JSON.stringify(response.data)
      );
    }

    rioters.value = response.data.map((rioter) => ({
      ...rioter,
      photo_name: rioter.photo_name, // ✅ Ensure this is preserved
      latitude: parseFloat(rioter.latitude),
      longitude: parseFloat(rioter.longitude),
    }));
  } catch (err) {
    console.error("Error fetching rioters:", err);
    error.value = `Failed to fetch rioters: ${err.message}`;
  } finally {
    loading.value = false;
  }
  mapKey.value++;
};

const getCoordinates = () => {
  return new Promise((resolve, reject) => {
    const options = {
      enableHighAccuracy: true,
      timeout: 3000, // Max 3 seconds wait
      maximumAge: 5 * 60 * 1000, // Cache for 5 minutes
    };

    navigator.geolocation.getCurrentPosition(
      (position) =>
        resolve({
          lat: position.coords.latitude,
          lng: position.coords.longitude,
        }),
      (error) => {
        console.error("Geolocation error:", error);
        reject(`Error: ${error.message}`);
      },
      options
    );
  });
};
</script>

<style>
@tailwind base;
@tailwind components;
@tailwind utilities;

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.map-move {
  transition: all 0.5s ease-out;
}
</style>

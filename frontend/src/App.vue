<template>
  <div id="app">
    <!-- Navigation Menu appears on every page -->
    <Navigation />

    <!-- Routed content will appear here -->
    <router-view />

    <!-- Additional content (for example, your sidebar, map, etc.) -->
    <div class="min-h-screen bg-gray-50 flex">
      <!-- Mobile Menu Button -->
      <button
        class="lg:hidden fixed top-4 right-4 z-50 p-2 bg-white rounded-lg shadow-md"
        @click="showMobileSidebar = !showMobileSidebar"
      >
        ☰
      </button>

      <!-- Sidebar -->
      <div
        class="w-96 bg-white border-r border-gray-200 flex flex-col fixed lg:fixed top-0 bottom-0 z-40"
        :class="{
          'translate-x-0': showMobileSidebar,
          '-translate-x-full lg:translate-x-0': !showMobileSidebar,
          'shadow-xl lg:shadow-none': showMobileSidebar,
        }"
      >
        <div class="h-full flex flex-col">
          <!-- Close Button for Mobile -->
          <button
            class="lg:hidden mb-4 text-gray-600 hover:text-gray-800"
            @click="showMobileSidebar = false"
          >
            ✕ Close
          </button>

          <!-- Fixed Header Section -->
          <div class="p-6">
            <h3 class="text-xl font-semibold text-gray-900 mb-4">
              Search The J6 Rioters Database
            </h3>
            <search-filters @filters-changed="handleFiltersChange" />
          </div>

          <!-- Scrollable Rioters List -->
          <div class="flex-1 overflow-y-auto px-6">
            <ul v-if="filteredRioters.length > 0" class="space-y-4">
              <li
                v-for="rioter in filteredRioters"
                :key="rioter.id"
                class="cursor-pointer p-4 hover:bg-gray-50 shadow rounded-lg"
                :class="{ 'bg-blue-50': selectedRioter?.id === rioter.id }"
                @click="selectRioter(rioter)"
              >
                <div class="flex items-center space-x-4">
                  <img
                    :src="getImageUrl(rioter.photo_name)"
                    class="h-12 w-12 rounded-full object-cover"
                    @error="handleImageError"
                  />
                  <div>
                    <h4 class="font-medium text-gray-900">
                      {{ rioter.first_name }} {{ rioter.last_name }}
                    </h4>
                    <p class="text-sm text-gray-500">
                      {{ [rioter.city, rioter.state].filter(Boolean).join(", ") }}
                    </p>
                  </div>
                </div>
              </li>
            </ul>

            <!-- No Results Message -->
            <div
              v-else-if="!loading"
              class="bg-white shadow rounded-lg p-6 text-center mt-6"
            >
              <p class="text-gray-500">No results found matching your filters.</p>
            </div>
          </div>
          <!-- <button
          v-if="currentPage < totalPages"
          @click="loadMoreRioters"
          class="mt-4 px-4 py-2 bg-blue-500 text-white rounded-lg w-full"
        >
          Load More
        </button> -->
          <!-- Fixed Footer Section -->
          <div class="p-6">
            <button
              class="mt-4 px-4 py-2 bg-blue-500 text-white rounded-lg w-full"
              @click="toggleFetchMode"
            >
              {{ fetchMode === "all" ? "Show Nearby" : "Show All" }}
            </button>
            <BasePagination
              v-if="fetchMode === 'all' && !currentFilters.state"
              :current-page="currentPage"
              :total-pages="totalPages"
              :page-size="pageSize"
              @page-changed="handlePageChange"
            />
          </div>
        </div>
      </div>

      <!-- Main Content (Map) -->
      <div class="flex-1 relative flex flex-col min-h-0" @click="closeSidebarOnMobile">
        <!-- Map Container -->
        <div class="sticky top-0 flex-1 min-h-0">
          <div class="h-full w-full relative">
            <!-- Loading Spinner -->
            <div
              v-if="loading"
              class="absolute inset-0 bg-gray-100/50 z-10 flex items-center justify-center"
            >
              <div
                class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"
              />
            </div>

            <!-- Map Component -->
            <rioters-map
              class="h-full w-full"
              :rioters="filteredRioters"
              :bounds="manualBounds || mapBounds"
              :selected-rioter="selectedRioter"
              @marker-click="selectRioter"
              :state-counts="stateCounts"
            />
          </div>
        </div>

        <!-- Flyout Details Panel -->
        <div
          v-if="selectedRioter"
          class="fixed inset-0 bg-black bg-opacity-50 lg:bg-transparent z-50"
          @mousedown.self="closePanel"
          @touchstart.self="closePanel"
        >
          <div
            class="absolute right-0 top-0 h-full w-full max-w-md bg-white shadow-xl lg:rounded-l-lg overflow-y-auto"
            role="dialog"
            aria-modal="true"
            aria-labelledby="panel-heading"
          >
            <div class="p-6">
              <button
                class="mb-4 text-gray-600 hover:text-gray-800"
                @click="selectedRioter = null"
              >
                ← Back to list
              </button>

              <!-- Detailed Rioter Content -->
              <div class="space-y-6">
                <div class="flex items-center space-x-4">
                  <img
                    :src="getImageUrl(selectedRioter.photo_name)"
                    class="h-32 w-32 rounded-full object-cover border-4 border-gray-200"
                    @error="handleImageError"
                  />
                  <div>
                    <h2 class="text-2xl font-bold text-gray-900">
                      {{ selectedRioter.first_name }}
                      {{ selectedRioter.last_name }}
                    </h2>
                    <p v-if="selectedRioter.age" class="text-gray-600">
                      Age: {{ selectedRioter.age }}
                    </p>
                  </div>
                </div>

                <!-- Location -->
                <div v-if="selectedRioter.city || selectedRioter.state" class="text-sm">
                  <h3 class="font-semibold">Location:</h3>
                  <p class="text-gray-600">
                    {{
                      [selectedRioter.city, selectedRioter.state]
                        .filter(Boolean)
                        .join(", ")
                    }}
                  </p>
                </div>

                <!-- Summary -->
                <div v-if="selectedRioter.summary" class="text-sm">
                  <h3 class="font-semibold">Summary:</h3>
                  <p class="mt-1 text-gray-600">
                    {{ selectedRioter.summary }}
                  </p>
                </div>

                <!-- Jurisdiction -->
                <div v-if="selectedRioter.jurisdiction" class="text-sm">
                  <h3 class="font-semibold">Jurisdiction:</h3>
                  <p class="text-gray-600">
                    {{ selectedRioter.jurisdiction }}
                  </p>
                </div>

                <!-- Charges -->
                <div v-if="selectedRioter.charges" class="text-sm">
                  <h3 class="font-semibold">Charges:</h3>
                  <p class="mt-1 text-gray-600">
                    {{ selectedRioter.charges }}
                  </p>
                </div>

                <!-- Case Status -->
                <div v-if="selectedRioter.case_status" class="text-sm">
                  <h3 class="font-semibold">Case Status:</h3>
                  <p class="mt-1 text-gray-600">
                    {{ selectedRioter.case_status }}
                  </p>
                </div>

                <!-- Case Updates -->
                <div v-if="selectedRioter.case_updates" class="text-sm">
                  <h3 class="font-semibold">Case Updates:</h3>
                  <p class="mt-1 text-gray-600">
                    {{ selectedRioter.case_updates }}
                  </p>
                </div>

                <!-- Tags -->
                <div class="flex flex-wrap gap-2">
                  <span
                    v-if="selectedRioter.violence_assault"
                    class="px-2 py-1 bg-red-100 text-red-800 text-xs rounded-full"
                  >
                    Violence/Assault
                  </span>
                  <span
                    v-if="selectedRioter.conspiracy"
                    class="px-2 py-1 bg-orange-100 text-orange-800 text-xs rounded-full"
                  >
                    Conspiracy
                  </span>
                  <span
                    v-if="selectedRioter.property"
                    class="px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded-full"
                  >
                    Property Damage
                  </span>
                  <span
                    v-if="selectedRioter.military_le"
                    class="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full"
                  >
                    Military/LE
                  </span>
                  <span
                    v-if="selectedRioter.extremist"
                    class="px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full"
                  >
                    Extremist
                  </span>
                  <span
                    v-if="selectedRioter.sentenced"
                    class="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full"
                  >
                    Sentenced
                  </span>
                </div>

                <!-- Charges Link -->
                <div v-if="selectedRioter.charges_link" class="mt-4">
                  <a
                    :href="selectedRioter.charges_link"
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
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from "vue";
import SearchFilters from "./components/SearchFilters.vue";
import RiotersMap from "./components/RiotersMap.vue";
import BasePagination from "./components/BasePagination.vue"; // Updated import
import api from "./api"; // Only one import
import Navigation from "./components/Navigation.vue";
// Component registration is automatic with <script setup>
// State
// const loadMoreRioters = () => {
//   if (currentPage.value < totalPages.value) {
//     currentPage.value += 1;
//     fetchRioters(true); // Append new data instead of replacing
//   }
// };
const stateCounts = ref({}); // Store state counts
// const selectedState = ref(""); // Stores the selected state
// const rioterCount = ref(null); // Stores the count of rioters in the selected state

const fetchStateCounts = async () => {
  try {
    const response = await api.get("/rioters/count-by-state");
    stateCounts.value = response.data;
  } catch (err) {
    console.error("Error fetching state counts:", err);
  }
};
const closeSidebarOnMobile = () => {
  if (window.innerWidth < 1024) {
    showMobileSidebar.value = false;
  }
};
const manualBounds = ref(null);
const showMobileSidebar = ref(false);
const selectedRioter = ref(null);
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
const fetchMode = ref("all");
const userLocation = ref(null);
const mapKey = ref(0);

// State
const currentPage = ref(1);
const pageSize = ref(1);
const totalPages = ref(1);
const totalItems = ref(0);

const closePanel = () => {
  selectedRioter.value = null;
  document.body.classList.remove("overflow-hidden");
};

const handleEsc = (e) => {
  if (e.key === "Escape" && selectedRioter.value) {
    closePanel();
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleEsc);
  fetchStateCounts();
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleEsc);
});

// Optional: Prevent background scroll when panel is open
watch(selectedRioter, (newVal) => {
  if (newVal) {
    document.body.classList.add("overflow-hidden");
  } else {
    document.body.classList.remove("overflow-hidden");
  }
});

// Computed Properties
const filteredRioters = computed(() => {
  return rioters.value.filter((rioter) => {
    const searchText = (currentFilters.value.searchText || "").toLowerCase();
    const stateFilter = (currentFilters.value.state || "").toLowerCase();
    const chargeType = currentFilters.value.charges;
    const statusFilter = (currentFilters.value.status || "").toLowerCase();
    const activeAffiliations = Object.entries(currentFilters.value.affiliations)
      .filter(([, value]) => value)
      .map(([key]) => key);

    // Search text filter
    if (
      searchText &&
      !(
        `${rioter.first_name} ${rioter.last_name}`?.toLowerCase().includes(searchText) ||
        rioter.summary?.toLowerCase().includes(searchText) ||
        rioter.charges?.toLowerCase().includes(searchText)
      )
    ) {
      return false;
    }

    // State filter
    if (stateFilter && rioter.state?.toLowerCase() !== stateFilter) {
      return false;
    }

    // Charge type filter
    if (chargeType && !rioter[chargeType]) return false;

    // Status filter
    if (statusFilter && !rioter.case_status?.toLowerCase().includes(statusFilter)) {
      return false;
    }

    // Affiliation filters
    if (
      activeAffiliations.length > 0 &&
      !activeAffiliations.every((affiliation) => rioter[affiliation])
    ) {
      return false;
    }

    return true;
  });
});

const mapBounds = computed(() => {
  const validRioters = filteredRioters.value.filter(
    (r) => r.latitude && r.longitude && !isNaN(r.latitude) && !isNaN(r.longitude)
  );

  if (validRioters.length === 0) {
    // Fallback to US bounds if no valid markers
    return [
      [-125.0, 24.0], // SW
      [-66.93457, 49.5904], // NE
    ];
  }
  const lngs = validRioters.map((r) => r.longitude);
  const lats = validRioters.map((r) => r.latitude);

  return [
    [Math.min(...lngs), Math.min(...lats)], // SW
    [Math.max(...lngs), Math.max(...lats)], // NE
  ];
});

// Methods
const toggleFetchMode = () => {
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

// Modify handleFiltersChange
const handleFiltersChange = (filters) => {
  // Reset pagination when a state filter is applied
  console.log("Applying filters:", filters);

  // If a state is selected, reset fetchMode to "all"
  if (filters.state) {
    fetchMode.value = "all"; // ✅ Ensure all rioters for the state are fetched
    currentPage.value = 1; // Reset to the first page
    pageSize.value = 50; // Adjust based on need
  }

  // Ensure the state parameter is properly set
  currentFilters.value = {
    searchText: filters.searchText || "",
    state: filters.state || "", // Ensure state is handled
    charges: filters.charges || "",
    status: filters.status || "",
    affiliations: { ...filters.affiliations }, // Clone affiliations object
  };
  // If any checkbox is selected, force fetchMode to "all"
  if (Object.values(filters.affiliations).some((val) => val)) {
    fetchMode.value = "all";
  }
  // If any dropdown or checkbox is applied, reset fetchMode to "all"
  if (
    filters.state ||
    filters.charges ||
    filters.status ||
    Object.values(filters.affiliations).some((val) => val)
  ) {
    fetchMode.value = "all";
  }

  // Update map bounds if state filter is applied
  // Update map bounds if state filter is applied
  if (filters.state) {
    const stateCenters = {
      alaska: [-152.404419, 64.200841], // Example coordinates for Alask
      al: [-86.902298, 32.318231], // Alabama
      ak: [-154.493062, 63.588753], // Alaska
      az: [-111.093731, 34.048927], // Arizona
      ar: [-92.373123, 34.969704], // Arkansas
      ca: [-119.417931, 37.184092], // California
      co: [-105.311104, 39.550051], // Colorado
      ct: [-72.755371, 41.597782], // Connecticut
      de: [-75.507141, 39.318523], // Delaware
      fl: [-81.515754, 27.994402], // Florida
      ga: [-83.357567, 32.165622], // Georgia
      hi: [-155.665857, 20.796667], // Hawaii
      id: [-114.742041, 44.068202], // Idaho
      il: [-89.398528, 40.633125], // Illinois
      in: [-86.134902, 40.267194], // Indiana
      ia: [-93.581543, 41.878003], // Iowa
      ks: [-98.484246, 38.498778], // Kansas
      ky: [-85.759407, 37.839333], // Kentucky
      la: [-92.145024, 31.244823], // Louisiana
      me: [-69.445469, 45.253783], // Maine
      md: [-76.641271, 39.045755], // Maryland
      ma: [-71.382437, 42.407211], // Massachusetts
      mi: [-85.602364, 44.314844], // Michigan
      mn: [-94.685899, 46.729553], // Minnesota
      ms: [-89.398528, 32.354668], // Mississippi
      mo: [-92.373123, 38.573936], // Missouri
      mt: [-110.362566, 46.879682], // Montana
      ne: [-99.901813, 41.492537], // Nebraska
      nv: [-116.419389, 38.80261], // Nevada
      nh: [-71.572395, 43.193852], // New Hampshire
      nj: [-74.405661, 40.058324], // New Jersey
      nm: [-105.870091, 34.51994], // New Mexico
      ny: [-75.144424, 43.156168], // New York
      nc: [-79.0193, 35.759573], // North Carolina
      nd: [-99.784012, 47.551493], // North Dakota
      oh: [-82.907123, 40.417287], // Ohio
      ok: [-97.092877, 35.007752], // Oklahoma
      or: [-120.554201, 44.045674], // Oregon
      pa: [-77.194525, 41.203322], // Pennsylvania
      ri: [-71.477429, 41.580095], // Rhode Island
      sc: [-81.163725, 33.836081], // South Carolina
      sd: [-99.438828, 43.969515], // South Dakota
      tn: [-86.580447, 35.517491], // Tennessee
      tx: [-99.359349, 31.816038], // Texas
      ut: [-111.950684, 39.41922], // Utah
      vt: [-72.577841, 44.558803], // Vermont
      va: [-78.656894, 37.431573], // Virginia
      wa: [-120.740139, 47.751074], // Washington
      wv: [-80.954456, 38.597626], // West Virginia
      wi: [-89.616508, 44.78444], // Wisconsin
      wy: [-107.290284, 43.075968], // Wyoming
    };
    const center = stateCenters[filters.state.toLowerCase()];
    if (center) {
      manualBounds.value = [
        [center[0] - 2, center[1] - 1],
        [center[0] + 2, center[1] + 1],
      ];
    }
  } else {
    manualBounds.value = null;
  }

  fetchRioters(); // Re-fetch data
};
const getCoordinates = () => {
  return new Promise((resolve, reject) => {
    navigator.geolocation.getCurrentPosition(
      (position) =>
        resolve({
          lat: position.coords.latitude,
          lng: position.coords.longitude,
        }),
      (error) => reject(error)
    );
  });
};

const fetchRioters = async (append = false) => {
  loading.value = true;
  error.value = null;

  try {
    let response;
    const params = {
      // page: currentPage.value,
      page: 1,
      page_size: 170, // Fetch more results per page if necessary
    };

    // if (currentFilters.value.state) {
    //   params.state = currentFilters.value.state;
    // }
    console.log("Fetching rioters with params:", params); // Debugging

    // response = await api.get("/rioters", { params });

    // Add non-empty filters to the params object
    Object.entries(currentFilters.value).forEach(([key, value]) => {
      if (value !== "" && value !== null && value !== undefined) {
        params[key] = value;
      }
    });
    // Ensure all states are fetched when no specific state is selected
    // if (!currentFilters.value.state) {
    //   delete params.state; // Remove state filter from query
    // }
    // ✅ Flatten affiliations object
    if (currentFilters.value.affiliations) {
      Object.entries(currentFilters.value.affiliations).forEach(([key, value]) => {
        if (value) {
          params[`affiliations.${key}`] = value; // Convert to API format
        }
      });
    }

    response = await api.get("/rioters", { params });

    // ✅ Append results when "Load More" is clicked
    rioters.value = append
      ? [...rioters.value, ...response.data.data] // Append new results
      : response.data.data; // Replace results when fetching fresh

    totalItems.value = response.data.total || response.data.length;
    totalPages.value = Math.ceil(totalItems.value / params.page_size);

    // Handle response
    // rioters.value = response.data.data || response.data;
    // totalItems.value = response.data.total || response.data.length;
    // totalPages.value = response.data.pages || 1;

    // Apply all active filters
    if (currentFilters.value.searchText) {
      params.searchText = currentFilters.value.searchText;
    }
    if (currentFilters.value.state) {
      response = await api.get("/rioters/by-state", {
        params: { state: currentFilters.value.state },
      });
      rioters.value = response.data; // Use the grouped response
    } else {
      // Fetch all rioters if no state filter is applied
      response = await api.get("/rioters", { params });
      rioters.value = response.data.data || response.data;
    }
    if (currentFilters.value.charges) {
      params.charges = currentFilters.value.charges;
    }
    if (currentFilters.value.status) {
      params.status = currentFilters.value.status;
    } // Flatten the affiliations object
    if (currentFilters.value.affiliations) {
      Object.entries(currentFilters.value.affiliations).forEach(([key, value]) => {
        if (value !== "" && value !== null && value !== undefined) {
          params[`affiliations.${key}`] = value; // Flatten to affiliations.military_le, etc.
        }
      });
      delete params.affiliations; // Remove the nested object
    }

    if (fetchMode.value === "all") {
      response = await api.get("/rioters", { params });
    } else {
      // Nearby mode logic
      let coords;
      try {
        coords = await getCoordinates();
        userLocation.value = coords;
      } catch (err) {
        userLocation.value = { lat: 34.052235, lng: -118.243683 };
      }

      const nearbyParams = {
        lng: coords.lng,
        lat: coords.lat,
        radius: 50000,
        ...params, // Use the flattened params
      };

      response = await api.get("/rioters/nearby", { params: nearbyParams });
    }

    // Handle response
    rioters.value = response.data.data || response.data;
    totalItems.value = response.data.total || response.data.length;
    totalPages.value = response.data.pages || 1;
  } catch (err) {
    console.error("Fetch error:", err);
    error.value = `Failed to fetch rioters: ${err.message}`;
  } finally {
    loading.value = false;
    mapKey.value++;
  }
};
// Add page change handler
const handlePageChange = (newPage) => {
  currentPage.value = newPage;
  fetchRioters();
};

const getImageUrl = (photoName) => {
  const baseUrl = "http://localhost:8080";
  return photoName?.trim()
    ? `${baseUrl}/photos/${encodeURIComponent(photoName)}`
    : `${baseUrl}/photos/placeholder.jpg`;
};

const handleImageError = (event) => {
  event.target.src = "http://localhost:8080/photos/placeholder.jpg";
};

const selectRioter = (rioter) => {
  selectedRioter.value = rioter; // Update the selected rioter
  showMobileSidebar.value = false; // Close sidebar on mobile
};

onMounted(fetchRioters);
</script>

<style>
@tailwind base;
@tailwind components;
@tailwind utilities;

@media (max-width: 1024px) {
  .lg\:rounded-l-lg {
    border-radius: 0;
  }
  .lg\:static {
    position: static;
  }
}

/* Custom Transitions */
.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s ease;
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(-100%);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

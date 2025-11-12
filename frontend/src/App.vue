<template>
  <div id="app">
    <Navigation />
    <router-view />
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
          <button
            class="lg:hidden mb-4 text-gray-600 hover:text-gray-800"
            @click="showMobileSidebar = false"
          >
            ✕ Close
          </button>
          <div class="p-2 flex-shrink-0 bg-white border-b border-gray-200">
            <h3 class="text-xl font-semibold text-gray-900">
              Search The J6 Rioters Database
            </h3>
            <!-- Count Display for Filters -->
            <div v-if="hasActiveFilters || fetchMode === 'nearby'" class="mb-2 p-3 bg-blue-50 border border-blue-200 rounded-md">
              <p class="text-lg font-bold text-blue-900">
                {{ filteredRioters.length || 0 }} {{ (filteredRioters.length || 0) === 1 ? 'Rioter' : 'Rioters' }} 
                <span v-if="fetchMode === 'nearby'">Nearby</span>
                <span v-else>Found</span>
              </p>
              <p class="text-xs text-blue-700">
                <span v-if="fetchMode === 'nearby'">Within 50km of your location</span>
                <span v-if="fetchMode === 'nearby' && hasActiveFilters"> • </span>
                <span v-if="currentFilters.state">State: {{ currentFilters.state }}</span>
                <!-- <span v-if="hasActiveCharges"> • Charges: {{ getChargeLabel(currentFilters.charges) }}</span> -->
                <span v-if="currentFilters.status"> • Status: {{ getStatusLabel(currentFilters.status) }}</span>
                <span v-if="hasActiveAffiliations"> • {{ getActiveAffiliationsLabel() }}</span>
                <span v-if="currentFilters.searchText"> • Search: "{{ currentFilters.searchText }}"</span>
              </p>
            </div>
            <search-filters 
              @filters-changed="handleFiltersChange" 
              @select-rioter="selectRioter" 
            />
          </div>
          <RiotersList
            :filteredRioters="filteredRioters"
            :selectedRioter="selectedRioter"
            :selectRioter="selectRioter"
            :loading="loading"
            :getImageUrl="getImageUrl"
            :handleImageError="handleImageError"
            :navigateToEdit="navigateToEdit"
          />
          <div
            class="p-2 bg-white border-t border-gray-200 flex flex-col items-center gap-2"
          >
            <button
              class="px-3 py-1 text-xs font-medium bg-blue-500 text-white rounded-md hover:bg-blue-600 transition"
              @click="toggleFetchMode"
            >
              {{ fetchMode === "all" ? "Show Nearby" : "Show All" }}
            </button>
            <div v-if="fetchMode === 'nearby'" class="text-center px-2">
              <p class="text-sm font-semibold text-gray-900">
                {{ rioters.length || 0 }} {{ (rioters.length || 0) === 1 ? 'Rioter' : 'Rioters' }} Found
              </p>
              <p class="text-xs text-gray-500">
                Within 50km of your location
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Content (Map) -->
      <div class="flex-1 relative flex flex-col min-h-0" @click="closeSidebarOnMobile">
        <div class="sticky top-0 flex-1 min-h-0">
          <div class="h-full w-full relative">
            <div
              v-if="loading"
              class="absolute inset-0 bg-gray-100/50 z-10 flex items-center justify-center"
            >
              <div
                class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"
              />
            </div>
            <rioters-map
              class="h-full w-full"
              :rioters="filteredRioters"
              :bounds="manualBounds || mapBounds"
              :selected-rioter="selectedRioter"
              ref="mapComponent"
              @marker-click="handleMarkerClick"
              @center-map="flyToMarker"
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
              <div class="space-y-6">
                <div class="flex items-center space-x-4">
                  <img
                    :src="getImageUrl(selectedRioter.photo_name)"
                    class="h-32 w-32 rounded-full object-cover border-4 border-gray-200"
                    @error="handleImageError"
                  />
                  <div>
                    <h2 class="text-2xl font-bold text-gray-900">
                      {{ selectedRioter.first_name }} {{ selectedRioter.last_name }}
                    </h2>
                    <p v-if="selectedRioter.age" class="text-gray-600">
                      Age: {{ selectedRioter.age }}
                    </p>
                  </div>
                </div>
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
                <div v-if="selectedRioter.summary" class="text-sm">
                  <h3 class="font-semibold">Summary:</h3>
                  <p class="mt-1 text-gray-600">{{ selectedRioter.summary }}</p>
                </div>
                <div v-if="selectedRioter.jurisdiction" class="text-sm">
                  <h3 class="font-semibold">Jurisdiction:</h3>
                  <p class="text-gray-600">{{ selectedRioter.jurisdiction }}</p>
                </div>
                <div v-if="selectedRioter.charges" class="text-sm">
                  <h3 class="font-semibold">Charges:</h3>
                  <p class="mt-1 text-gray-600">{{ selectedRioter.charges }}</p>
                </div>
                <div v-if="selectedRioter.case_status" class="text-sm">
                  <h3 class="font-semibold">Case Status:</h3>
                  <p class="mt-1 text-gray-600">{{ selectedRioter.case_status }}</p>
                </div>
                <div v-if="selectedRioter.case_updates" class="text-sm">
                  <h3 class="font-semibold">Case Updates:</h3>
                  <p class="mt-1 text-gray-600">{{ selectedRioter.case_updates }}</p>
                </div>
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
import { useRouter } from "vue-router";
import SearchFilters from "./components/SearchFilters.vue";
import RiotersMap from "./components/RiotersMap.vue";
import Navigation from "./components/Navigation.vue";
import RiotersList from "./components/RiotersList.vue";
import api from "./api";

const mapComponent = ref(null);
const router = useRouter();

const rioters = ref([]);
const loading = ref(false);
const error = ref(null);
const selectedRioter = ref(null);
const fetchMode = ref("all");
const manualBounds = ref(null);
const currentFilters = ref({
  searchText: "",
  state: "",
  charges: "",
  status: "",
  affiliations: {
    military_le: false,
    extremist: false,
    sentenced: false,
    commuted: false,
  },
});
const currentPage = ref(1);
const pageSize = ref(50);
const totalItems = ref(0);
const totalPages = ref(1);
const showMobileSidebar = ref(false);

const filteredRioters = computed(() => {
  // Ensure we always return an array
  if (!Array.isArray(rioters.value)) {
    return [];
  }
  return rioters.value;
});

// Check if any filters are active
const hasActiveFilters = computed(() => {
  // Charges filter commented out - not working correctly
  // const hasCharges = typeof currentFilters.value.charges === 'string' 
  //   ? !!currentFilters.value.charges
  //   : Object.values(currentFilters.value.charges || {}).some(v => v);
  
  return !!(
    currentFilters.value.searchText ||
    currentFilters.value.state ||
    // hasCharges ||
    currentFilters.value.status ||
    Object.values(currentFilters.value.affiliations || {}).some(v => v)
  );
});

// Check if any affiliations are active
const hasActiveAffiliations = computed(() => {
  return Object.values(currentFilters.value.affiliations || {}).some(v => v);
});

// Check if any charges are active - COMMENTED OUT - Not working correctly
// const hasActiveCharges = computed(() => {
//   if (typeof currentFilters.value.charges === 'string') {
//     return !!currentFilters.value.charges;
//   }
//   return Object.values(currentFilters.value.charges || {}).some(v => v);
// });

// Helper functions for labels - Charges filter commented out
// const getChargeLabel = (charges) => {
//   if (typeof charges === 'string') {
//     // Legacy support
//     const labels = {
//       'violence_assault': 'Violence/Assault',
//       'conspiracy': 'Conspiracy',
//       'property': 'Property Destruction'
//     };
//     return labels[charges] || charges;
//   }
//   // New format: object with multiple charge types
//   const active = [];
//   if (charges?.violence_assault) active.push('Violence/Assault');
//   if (charges?.conspiracy) active.push('Conspiracy');
//   if (charges?.property) active.push('Property');
//   return active.join(', ') || 'Any';
// };

const getStatusLabel = (status) => {
  const labels = {
    'pleaded-not-guilty': 'Pleaded Not Guilty',
    'pleaded-guilty': 'Pleaded Guilty',
    'acquitted': 'Acquitted',
    'convicted': 'Convicted',
    'dismissed': 'Dismissed'
  };
  return labels[status] || status;
};

const getActiveAffiliationsLabel = () => {
  const active = [];
  if (currentFilters.value.affiliations?.military_le) active.push('Military/LEO');
  if (currentFilters.value.affiliations?.extremist) active.push('Extremist');
  if (currentFilters.value.affiliations?.sentenced) active.push('Sentenced');
  if (currentFilters.value.affiliations?.commuted) active.push('Commuted');
  return active.join(', ');
};

const mapBounds = computed(() => {
  // Ensure filteredRioters is an array before filtering
  if (!Array.isArray(filteredRioters.value)) {
    return [
      [-125.0, 24.0], // SW
      [-66.93457, 49.5904], // NE
    ];
  }
  const validRioters = filteredRioters.value.filter(
    (r) => r.latitude && r.longitude && !isNaN(r.latitude) && !isNaN(r.longitude)
  );
  if (validRioters.length === 0) {
    return [
      [-125.0, 24.0], // SW
      [-66.93457, 49.5904], // NE
    ];
  }
  const lngs = validRioters.map((r) => r.longitude);
  const lats = validRioters.map((r) => r.latitude);
  const padding = 0.5;
  return [
    [Math.min(...lngs) - padding, Math.min(...lats) - padding], // SW
    [Math.max(...lngs) + padding, Math.max(...lats) + padding], // NE
  ];
});

const fetchRioters = async () => {
  loading.value = true;
  error.value = null;
  try {
    if (fetchMode.value === "nearby") {
      // Get user's actual location via geolocation API - this is the primary source
      let lat, lng;
      let locationSource = "unknown";
      
      // Try geolocation first (user's actual location)
      if (navigator.geolocation) {
        try {
          const position = await new Promise((resolve, reject) => {
            navigator.geolocation.getCurrentPosition(
              resolve, 
              reject, 
              {
                timeout: 10000, // 10 seconds
                maximumAge: 300000, // 5 minutes - use cached location if recent
                enableHighAccuracy: true // Get best accuracy for "nearby" feature
              }
            );
          });
          lat = position.coords.latitude;
          lng = position.coords.longitude;
          locationSource = "your location";
          console.log("📍 Using your browser location:", lat, lng);
        } catch (geoError) {
          console.warn("⚠️ Geolocation failed:", geoError.message);
          
          // Fallback to map center if geolocation fails
          let center = null;
          for (let i = 0; i < 10; i++) {
            await new Promise(resolve => setTimeout(resolve, 100));
            if (mapComponent.value?.getMapCenter) {
              center = mapComponent.value.getMapCenter();
              if (center && typeof center.lat === 'number' && typeof center.lng === 'number' && 
                  !isNaN(center.lat) && !isNaN(center.lng)) {
                lat = center.lat;
                lng = center.lng;
                locationSource = "map center (geolocation unavailable)";
                console.log("📍 Using map center as fallback:", lat, lng);
                break;
              }
            }
          }
          
          // Final fallback to default center
          if (!lat || !lng) {
            lat = 39.8283;
            lng = -98.5795;
            locationSource = "default (US center)";
            console.log("📍 Using default center:", lat, lng);
          }
        }
      } else {
        // Geolocation not supported - use map center
        console.warn("⚠️ Geolocation not supported by browser");
        let center = null;
        for (let i = 0; i < 10; i++) {
          await new Promise(resolve => setTimeout(resolve, 100));
          if (mapComponent.value?.getMapCenter) {
            center = mapComponent.value.getMapCenter();
            if (center && typeof center.lat === 'number' && typeof center.lng === 'number' && 
                !isNaN(center.lat) && !isNaN(center.lng)) {
              lat = center.lat;
              lng = center.lng;
              locationSource = "map center (geolocation not supported)";
              console.log("📍 Using map center:", lat, lng);
              break;
            }
          }
        }
        
        // Final fallback
        if (!lat || !lng) {
          lat = 39.8283;
          lng = -98.5795;
          locationSource = "default (US center)";
          console.log("📍 Using default center:", lat, lng);
        }
      }
      
      // Validate coordinates
      if (!lat || !lng || isNaN(lat) || isNaN(lng) || Math.abs(lat) > 90 || Math.abs(lng) > 180) {
        console.error("Invalid coordinates:", lat, lng);
        error.value = "Invalid location coordinates";
        return;
      }
      
            // Fetch nearby rioters (50km radius) with filters applied
            console.log(`🔍 Fetching nearby rioters for ${locationSource}:`, lat, lng, "radius: 50km");
            const nearbyParams = {
              lat: lat,
              lng: lng,
              radius: 50000 // 50km in meters
            };
            
            // Apply current filters to nearby search
            if (currentFilters.value.state) {
              nearbyParams.state = currentFilters.value.state;
            }
            if (currentFilters.value.status) {
              nearbyParams.status = currentFilters.value.status;
            }
            if (currentFilters.value.searchText) {
              nearbyParams.searchText = currentFilters.value.searchText;
            }
            // Apply affiliation filters
            if (currentFilters.value.affiliations) {
              Object.entries(currentFilters.value.affiliations).forEach(([key, value]) => {
                if (value) {
                  nearbyParams[key] = 'true';
                }
              });
            }
            
            const response = await api.get("/rioters/nearby", {
              params: nearbyParams
            });
            console.log("✅ Nearby response:", response.data);
            rioters.value = Array.isArray(response.data) ? response.data : [];
            totalItems.value = rioters.value.length;
            totalPages.value = 1;
            console.log(`✅ Found ${rioters.value.length} nearby rioters (using ${locationSource}) with filters applied`);
            
            // Show user which location was used
            if (rioters.value.length === 0) {
              const filterInfo = hasActiveFilters.value ? " (with current filters)" : "";
              error.value = `No rioters found within 50km of ${locationSource === "map center" ? "current map view" : locationSource}${filterInfo}. Try panning the map to a different area or adjusting filters.`;
            }
    } else {
      // Regular fetch with filters
      // If no filters, fetch ALL rioters for map display (in batches of 1000)
      if (!hasActiveFilters.value && currentPage.value === 1) {
        console.log("🗺️ No filters applied - fetching all rioters for map display");
        const allRioters = [];
        let page = 1;
        const batchSize = 1000; // Backend max limit
        let hasMore = true;
        
        while (hasMore) {
          const params = {
            page: page,
            page_size: batchSize,
            ...currentFilters.value,
          };
          if (currentFilters.value.affiliations) {
            Object.entries(currentFilters.value.affiliations).forEach(([key, value]) => {
              params[key] = value;
            });
            delete params.affiliations;
          }
          
          const response = await api.get("/rioters", { params });
          const batch = response.data.data || response.data;
          allRioters.push(...batch);
          
          totalItems.value = response.data.total;
          
          // Check if we've fetched all
          if (allRioters.length >= totalItems.value || batch.length < batchSize) {
            hasMore = false;
          } else {
            page++;
          }
        }
        
        rioters.value = allRioters;
        totalPages.value = Math.ceil(totalItems.value / pageSize.value);
        console.log(`📊 Fetched ALL ${rioters.value.length} rioters for map display`);
      } else {
        // Use normal pagination when filters are applied
        // If state filter is active, fetch ALL rioters for that state (for map display)
        if (currentFilters.value.state && currentPage.value === 1) {
          console.log(`🗺️ State filter active (${currentFilters.value.state}) - fetching all rioters for map display`);
          const allRioters = [];
          let page = 1;
          const batchSize = 1000; // Backend max limit
          let hasMore = true;
          
          while (hasMore) {
            const params = {
              page: page,
              page_size: batchSize,
              ...currentFilters.value,
            };
            // Handle charges object - convert to individual query params (COMMENTED OUT - Not working correctly)
            // if (currentFilters.value.charges && typeof currentFilters.value.charges === 'object') {
            //   Object.entries(currentFilters.value.charges).forEach(([key, value]) => {
            //     if (value) {
            //       params[key] = 'true';
            //     }
            //   });
            //   delete params.charges;
            // }
            // Handle affiliations
            if (currentFilters.value.affiliations) {
              Object.entries(currentFilters.value.affiliations).forEach(([key, value]) => {
                if (value) {
                  params[key] = 'true';
                }
              });
              delete params.affiliations;
            }
            
            const response = await api.get("/rioters", { params });
            const batch = response.data.data || response.data;
            allRioters.push(...batch);
            
            totalItems.value = response.data.total;
            
            // Check if we've fetched all
            if (allRioters.length >= totalItems.value || batch.length < batchSize) {
              hasMore = false;
            } else {
              page++;
            }
          }
          
          rioters.value = allRioters;
          totalPages.value = Math.ceil(totalItems.value / pageSize.value);
          console.log(`📊 Fetched ALL ${rioters.value.length} rioters for ${currentFilters.value.state}`);
        } else {
          // Use normal pagination for other filters or when not on first page
          const params = {
            page: currentPage.value,
            page_size: pageSize.value,
            ...currentFilters.value,
          };
          // Handle charges object - convert to individual query params
          if (currentFilters.value.charges && typeof currentFilters.value.charges === 'object') {
            Object.entries(currentFilters.value.charges).forEach(([key, value]) => {
              if (value) {
                params[key] = 'true';
              }
            });
            delete params.charges;
          }
          // Handle affiliations
          if (currentFilters.value.affiliations) {
            Object.entries(currentFilters.value.affiliations).forEach(([key, value]) => {
              if (value) {
                params[key] = 'true';
              }
            });
            delete params.affiliations;
          }
          const response = await api.get("/rioters", { params });
          // Ensure we always assign an array
          const data = response.data.data || response.data;
          rioters.value = Array.isArray(data) ? data : [];
          totalItems.value = response.data.total;
          totalPages.value = Math.ceil(totalItems.value / pageSize.value);
          console.log(`📊 Fetched ${rioters.value.length} rioters (total available: ${totalItems.value})`);
        }
      }
    }
  } catch (err) {
    console.error("Fetch error:", err);
    error.value = `Failed to fetch rioters: ${err.message}`;
  } finally {
    loading.value = false;
  }
};

const handleFiltersChange = (filters) => {
  currentFilters.value = { ...filters };
  currentPage.value = 1;
  fetchRioters();
};

const toggleFetchMode = () => {
  // Don't clear filters - preserve them so they can overlap with nearby mode
  currentPage.value = 1;
  fetchMode.value = fetchMode.value === "all" ? "nearby" : "all";
  fetchRioters();
};

const photoUrl = "http://localhost:8085";
const getImageUrl = (photoName) =>
  photoName?.trim()
    ? `${photoUrl}/photos/${encodeURIComponent(photoName)}`
    : `${photoUrl}/photos/placeholder.jpg`;

const handleImageError = (event) => {
  event.target.src = `${photoUrl}/photos/placeholder.jpg`;
};

const selectRioter = async (rioter) => {
  if (fetchMode.value === "nearby") {
    try {
      const response = await api.get(`/rioters/${rioter.id}`);
      selectedRioter.value = response.data;
    } catch (error) {
      console.error("Error fetching rioter details:", error);
      selectedRioter.value = rioter; // Fallback to passed rioter
    }
  } else {
    selectedRioter.value = rioter;
  }
  showMobileSidebar.value = false;
  const rioterElement = document.querySelector(`[data-rioter-id="${rioter.id}"]`);
  if (rioterElement) {
    rioterElement.scrollIntoView({ behavior: "smooth", block: "center" });
  }
};

const navigateToEdit = (rioter) => {
  router.push(`/rioter/${rioter.id}/edit`);
};

const flyToMarker = (rioter) => {
  if (mapComponent.value?.flyToMarker) {
    mapComponent.value.flyToMarker(rioter);
  }
};

const handleMarkerClick = (rioter) => {
  selectRioter(rioter);
  flyToMarker(rioter);
};

const closePanel = () => {
  selectedRioter.value = null;
  document.body.classList.remove("overflow-hidden");
};

onMounted(() => {
  fetchRioters();
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && selectedRioter.value) {
      closePanel();
    }
  });
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", () => {});
});

watch(selectedRioter, (newVal) => {
  if (newVal) {
    document.body.classList.add("overflow-hidden");
  } else {
    document.body.classList.remove("overflow-hidden");
  }
});

watch(filteredRioters, () => {
  if (mapComponent.value && mapComponent.value.fitBounds) {
    mapComponent.value.fitBounds(mapBounds.value);
  }
});

const closeSidebarOnMobile = () => {
  if (window.innerWidth < 1024) {
    showMobileSidebar.value = false;
  }
};
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

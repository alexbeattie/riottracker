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
            <search-filters @filters-changed="handleFiltersChange" />
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
            class="p-2 bg-white border-t border-gray-200 flex items-center justify-center"
          >
            <button
              class="px-3 py-1 text-xs font-medium bg-blue-500 text-white rounded-md hover:bg-blue-600 transition"
              @click="toggleFetchMode"
            >
              {{ fetchMode === "all" ? "Show Nearby" : "Show All" }}
            </button>
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

const filteredRioters = computed(() => rioters.value);

const mapBounds = computed(() => {
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
    const params = {
      page: currentPage.value,
      page_size: currentFilters.value.state ? 1000 : pageSize.value,
      ...currentFilters.value,
    };
    if (currentFilters.value.affiliations) {
      Object.entries(currentFilters.value.affiliations).forEach(([key, value]) => {
        params[key] = value;
      });
      delete params.affiliations;
    }
    const response = await api.get("/rioters", { params });
    rioters.value = response.data.data || response.data;
    totalItems.value = response.data.total;
    totalPages.value = Math.ceil(totalItems.value / pageSize.value);
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
  currentFilters.value = {
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
  };
  fetchMode.value = fetchMode.value === "all" ? "nearby" : "all";
  fetchRioters();
};

const photoUrl = "http://localhost:8080";
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

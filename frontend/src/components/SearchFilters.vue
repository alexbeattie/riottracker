<template>
  <div class="bg-white shadow rounded-lg p-6 mb-2">
    <!-- Search & Reset Button (Inline) -->
    <div class="flex items-center gap-4 mb-4">
      <!-- 🔍 Search Box -->
      <div class="relative flex-1">
        <input
          v-model="filters.searchText"
          type="text"
          placeholder="Search by name, location, charges..."
          class="w-full px-4 py-2 pl-10 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          @input="debounceSearch"
          @change="emitFilters"
        />
        <span class="absolute left-3 top-2.5 text-gray-400">🔍</span>
      </div>

      <!-- Reset Filters Button -->
      <button
        class="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 transition"
        @click="resetFilters"
      >
        Reset Filters
      </button>
    </div>

    <!-- Dropdowns Grid (Inline) -->
    <div class="grid grid-cols-3 gap-4 mb-4 text-sm">
      <!-- State Dropdown -->
      <div class="relative">
        <label class="block text-gray-700 mb-1">State</label>
        <select
          v-model="filters.state"
          class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
          @change="emitFilters"
        >
          <option value="">All states...</option>
          <option v-for="state in states" :key="state" :value="state">
            {{ state }}
          </option>
        </select>
      </div>

      <!-- Charges Dropdown -->
      <div class="relative">
        <label class="block text-gray-700 mb-1">Charges</label>
        <select
          v-model="filters.charges"
          class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
          @change="emitFilters"
        >
          <option value="">Any charges...</option>
          <option value="violence_assault">Violence/assault</option>
          <option value="conspiracy">Conspiracy</option>
          <option value="property">Property destruction</option>
        </select>
      </div>

      <!-- Status Dropdown -->
      <div class="relative">
        <label class="block text-gray-700 mb-1 truncate">Case Status</label>
        <select
          v-model="filters.status"
          class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
          @change="emitFilters"
        >
          <option value="">Any case status...</option>
          <option value="pleaded-not-guilty">Pleaded not guilty</option>
          <option value="pleaded-guilty">Pleaded guilty to one or more charges</option>
          <option value="acquitted">Acquitted at trial</option>
          <option value="convicted">Convicted at trial</option>
          <option value="dismissed">Dismissed</option>
        </select>
      </div>
    </div>
    <!-- Display State Count -->
    <div v-if="rioterCount !== null" class="mb-4 p-3 bg-blue-100 text-blue-800 rounded">
      There are {{ rioterCount }} rioters in {{ selectedState }}.
    </div>
    <!-- Affiliation Filters (Inline) -->
    <div class="mt-4">
      <label class="block text-gray-700 mb-2">Affiliations</label>
      <div class="flex flex-wrap items-center gap-1">
        <button
          v-for="(label, key) in affiliationOptions"
          :key="key"
          class="px-2 py-1 rounded-lg text-xs font-medium transition whitespace-nowrap border border-gray-300 shadow-sm"
          :class="
            filters.affiliations[key]
              ? 'bg-blue-500 text-white border-blue-500'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          "
          @click="toggleAffiliation(key)"
        >
          {{ label }}
        </button>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, watch, defineEmits, onMounted } from "vue";
import api from "../api"; // Ensure this is your API file

// Define the emit function
const emit = defineEmits(["filters-changed"]);
const debounceTimeout = ref(null);
const loading = ref(false); // Add loading state
const stateCounts = ref({}); // Stores the count of rioters in each state
const selectedState = ref(""); // Tracks selected state
const rioterCount = ref(null); // Stores rioter count for selected state

const filters = ref({
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

const states = ref([
  "Alabama",
  "Alaska",
  "Arizona",
  "Arkansas",
  "California",
  "Colorado",
  "Connecticut",
  "D.C.",
  "Delaware",
  "Florida",
  "Georgia",
  "Hawaii",
  "Idaho",
  "Illinois",
  "Indiana",
  "Iowa",
  "Kansas",
  "Kentucky",
  "Louisiana",
  "Maine",
  "Maryland",
  "Massachusetts",
  "Michigan",
  "Minnesota",
  "Mississippi",
  "Missouri",
  "Montana",
  "Nebraska",
  "Nevada",
  "New Hampshire",
  "New Jersey",
  "New Mexico",
  "New York",
  "North Carolina",
  "North Dakota",
  "Ohio",
  "Oklahoma",
  "Oregon",
  "Pennsylvania",
  "Rhode Island",
  "South Carolina",
  "South Dakota",
  "Tennessee",
  "Texas",
  "Utah",
  "Vermont",
  "Virginia",
  "Washington",
  "West Virginia",
  "Wisconsin",
  "Wyoming",
]);

const affiliationOptions = ref({
  military_le: "Military/LEO",
  extremist: "Extremist Orgs",
  sentenced: "Sentenced",
});
// Fetch the state counts from the backend
const fetchStateCounts = async () => {
  try {
    const response = await api.get("/rioters/count-by-state");
    stateCounts.value = response.data;
  } catch (error) {
    console.error("Error fetching state counts:", error);
  }
};

const emitFilters = () => {
  loading.value = true; // Show loading indicator
  setTimeout(() => {
    const payload = {
      searchText: filters.value.searchText.trim(),
      state: filters.value.state,
      charges: filters.value.charges,
      status: filters.value.status,
      affiliations: { ...filters.value.affiliations },
    };
    console.log("Emitting filters:", payload);
    emit("filters-changed", payload);
    loading.value = false; // Hide loading indicator
  }, 500);
};

// Toggle affiliations (Button Click)
const toggleAffiliation = (key) => {
  filters.value.affiliations[key] = !filters.value.affiliations[key];
  emitFilters();
};

// Reset Filters
const resetFilters = async () => {
  filters.value = {
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

  selectedState.value = "All States"; // Reset display state text

  await fetchStateCounts(); // Fetch updated counts from the API

  rioterCount.value = stateCounts.value["total"] || 1567; // Default total rioters

  emitFilters(); // Trigger filter update
};
// Debounce search input
const debounceSearch = () => {
  clearTimeout(debounceTimeout.value);
  if (!filters.value.searchText.trim()) {
    emitFilters();
    return;
  }
  debounceTimeout.value = setTimeout(() => {
    emitFilters();
  }, 300);
};

// Watch for filter changes
watch(
  () => filters.value.state,
  (newState) => {
    if (!newState) {
      selectedState.value = "All States";
      rioterCount.value = stateCounts.value["total"] || 1567;
    } else {
      selectedState.value = newState;
      rioterCount.value = stateCounts.value[newState] || 0;
    }
    emitFilters();
  }
);
onMounted(fetchStateCounts);
</script>

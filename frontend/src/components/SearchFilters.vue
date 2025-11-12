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
          @focus="showSuggestions = true"
          @blur="hideSuggestionsWithDelay"
        />
        <span class="absolute left-3 top-2.5 text-gray-400">🔍</span>

        <!-- 🔽 Suggestions Dropdown -->
        <ul
          v-if="showSuggestions && searchSuggestions.length"
          class="absolute left-0 w-full bg-white border border-gray-300 rounded-md shadow-lg mt-1 z-50"
        >
          <li
            v-for="(suggestion, index) in searchSuggestions"
            :key="index"
            class="p-2 hover:bg-blue-100 cursor-pointer"
            @mousedown.prevent="selectSuggestion(suggestion)"
          >
            <span v-html="suggestion"></span>
            <!-- Highlighted results -->
          </li>
        </ul>
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
    <div class="grid grid-cols-2 gap-4 mb-4 text-sm">
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

      <!-- Charges - Multiple Selection Buttons (COMMENTED OUT - Not working correctly) -->
      <!--
      <div class="relative">
        <label class="block text-gray-700 mb-1">Charges</label>
        <div class="flex flex-wrap gap-1">
          <button
            v-for="(label, key) in chargeOptions"
            :key="key"
            class="px-2 py-1 rounded-lg text-xs font-medium transition whitespace-nowrap border border-gray-300 shadow-sm"
            :class="
              filters.charges[key]
                ? 'bg-red-500 text-white border-red-500'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            "
            @click="toggleCharge(key)"
          >
            {{ label }}
          </button>
        </div>
      </div>
      -->

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
const emit = defineEmits(["filters-changed", "select-rioter"]);
const debounceTimeout = ref(null);
const showSuggestions = ref(false); // Controls dropdown visibility
const searchSuggestions = ref([]); // Stores API search results

const loading = ref(false); // Add loading state
const stateCounts = ref({}); // Stores the count of rioters in each state
const selectedState = ref(""); // Tracks selected state
const rioterCount = ref(null); // Stores rioter count for selected state

const filters = ref({
  searchText: "",
  state: "",
  charges: {
    violence_assault: false,
    conspiracy: false,
    property: false,
  },
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

// Charges filter commented out - not working correctly
// const chargeOptions = ref({
//   violence_assault: "Violence/Assault",
//   conspiracy: "Conspiracy",
//   property: "Property",
// });
// Fetch suggestions from backend
const fetchSearchSuggestions = async () => {
  if (!filters.value.searchText.trim()) {
    searchSuggestions.value = [];
    return;
  }

  try {
    const response = await api.get(
      `/search/suggestions?term=${encodeURIComponent(filters.value.searchText)}`
    );
    searchSuggestions.value = response.data;
  } catch (error) {
    console.error("Error fetching search suggestions:", error);
  }
};
// Apply the selected suggestion
const selectSuggestion = async (suggestion) => {
  console.log("Clicked suggestion:", suggestion); // Debugging

  // Store the suggestion temporarily
  const plainName = suggestion.replace(/<\/?mark>/g, ""); // Remove HTML highlighting
  
  // Clear the search field temporarily to avoid automatic search
  const originalSearchText = filters.value.searchText;
  filters.value.searchText = "";
  searchSuggestions.value = [];
  showSuggestions.value = false;
  
  try {
    // Make a direct API call for this specific name
    const nameQuery = plainName.trim();
    console.log("Searching for exact match:", nameQuery);
    
    // Split the name into first and last name parts
    const nameParts = nameQuery.split(' ');
    let firstName = '';
    let lastName = '';
    
    if (nameParts.length >= 2) {
      firstName = nameParts[0];
      lastName = nameParts.slice(1).join(' ');
    } else {
      // If only one word, use it as both first and last name in the search
      firstName = lastName = nameParts[0];
    }
    
    // Perform an exact search using a custom parameter
    const exactResponse = await api.get(`/rioters`, {
      params: {
        page: 1,
        page_size: 50,
        search_exact: true,
        first_name: firstName,
        last_name: lastName
      }
    });
    
    console.log("Search response:", exactResponse.data);
    
    if (exactResponse.data.data && exactResponse.data.data.length > 0) {
      // First try to find someone with exactly this name
      const exactMatch = exactResponse.data.data.find(rioter => 
        `${rioter.first_name} ${rioter.last_name}`.toLowerCase() === nameQuery.toLowerCase()
      );
      
      if (exactMatch) {
        console.log("Found exact name match:", exactMatch);
        // Set the search text to show what was searched
        filters.value.searchText = plainName;
        // Emit both events
        emit("filters-changed", filters.value);
        emit("select-rioter", exactMatch);
        return;
      } 
      
      // If no exact match, use the first result as a fallback
      console.log("No exact match, using first result:", exactResponse.data.data[0]);
      filters.value.searchText = plainName;
      emit("filters-changed", filters.value);
      emit("select-rioter", exactResponse.data.data[0]);
    } else {
      // No results found, restore original search text
      console.log("No results found for:", nameQuery);
      filters.value.searchText = plainName;
      emit("filters-changed", filters.value);
    }
  } catch (error) {
    console.error("Error finding rioter by name:", error);
    // Restore original search on error
    filters.value.searchText = originalSearchText;
    emit("filters-changed", filters.value);
  }
};

// Fetch the state counts from the backend
const fetchStateCounts = async () => {
  try {
    const response = await api.get("/rioters/count-by-state");
    stateCounts.value = response.data;
  } catch (error) {
    console.error("Error fetching state counts:", error);
  }
};
// watch(filteredRioters, () => {
//   // Optional: Force map to re-center if needed
//   if (mapComponent.value && mapComponent.value.fitBounds) {
//     mapComponent.value.fitBounds(mapBounds.value);
//   }
// });
const emitFilters = () => {
  loading.value = true; // Show loading indicator
  setTimeout(() => {
    showSuggestions.value = false; // Hide dropdown
    const payload = {
      searchText: filters.value.searchText.trim(),
      state: filters.value.state,
      charges: filters.value.charges, // Now an object with multiple charge types
      status: filters.value.status,
      affiliations: { ...filters.value.affiliations },
    };
    console.log("Emitting filters:", payload);
    emit("filters-changed", payload);
    loading.value = false; // Hide loading indicator
  }, 500);
};

// Toggle charge types (Button Click) - COMMENTED OUT - Not working correctly
// const toggleCharge = (key) => {
//   filters.value.charges[key] = !filters.value.charges[key];
//   emitFilters();
// };

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
    charges: {
      violence_assault: false,
      conspiracy: false,
      property: false,
    },
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
    searchSuggestions.value = [];
    emitFilters();
    return;
  }
  
  // Make sure suggestions are visible during typing
  showSuggestions.value = true;
  
  // Immediately fetch suggestions as user types
  fetchSearchSuggestions();
  
  // Still use debounce for search to avoid too many API calls
  debounceTimeout.value = setTimeout(() => {
    emitFilters();
  }, 300);
};

// Hide suggestions after a short delay to allow click selection
const hideSuggestionsWithDelay = () => {
  setTimeout(() => {
    showSuggestions.value = false;
  }, 200);
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

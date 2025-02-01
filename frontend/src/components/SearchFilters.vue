<!-- src/components/SearchFilters.vue -->
<template>
  <div class="bg-white shadow rounded-lg p-6 mb-8">
    <!-- <h3 class="text-xl font-semibold text-gray-900 mb-6">Search The Database</h3> -->

    <!-- Text Search -->
    <div class="mb-4">
      <input
        v-model="filters.searchText"
        type="text"
        placeholder="Search anything..."
        class="w-full p-2 pl-8 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        @input="debounceSearch"
      />
    </div>

    <!-- Dropdowns Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <!-- State Dropdown -->
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

      <!-- Charges Dropdown -->
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

      <!-- Status Dropdown -->
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
        <option value="missing">Missing</option>
      </select>
    </div>

    <!-- Checkboxes -->
    <div class="space-y-4">
      <!-- New/Updated Checkbox -->
      <!-- <div class="flex items-center">
        <input
          type="checkbox"
          id="new"
          v-model="filters.isNewOrUpdated"
          @change="emitFilters"
          class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
        />
        <label for="new" class="ml-2 text-sm text-gray-700">New/updated</label>
      </div> -->

      <!-- Affiliations -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="(label, key) in affiliationOptions"
          :key="key"
          class="flex items-center"
        >
          <input
            :id="key"
            v-model="filters.affiliations[key]"
            type="checkbox"
            class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            @change="emitFilters"
          />
          <label :for="key" class="ml-2 text-sm text-gray-700">{{ label }}</label>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "SearchFilters",
  data() {
    return {
      debounceTimeout: null,
      filters: {
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
      },
      states: [
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
      ],
      stateCounts: {}, // Holds the number of rioters per state

      affiliationOptions: {
        military_le: "Military/law enforcement ties",
        extremist: "Ties to extremist or fringe groups",
        sentenced: "Sentenced",
        commuted: "Commuted",
      },
    };
  },
  watch: {
    // Optimized watchers with conditional emission
    "filters.state"(newVal, oldVal) {
      if (newVal !== oldVal) this.emitFilters();
    },
    "filters.charges"(newVal, oldVal) {
      if (newVal !== oldVal) this.emitFilters();
    },
    "filters.status"(newVal, oldVal) {
      if (newVal !== oldVal) this.emitFilters();
    },
    "filters.affiliations": {
      handler(newVal, oldVal) {
        if (JSON.stringify(newVal) !== JSON.stringify(oldVal)) {
          this.emitFilters();
        }
      },
      deep: true,
    },
  },
  beforeUnmount() {
    clearTimeout(this.debounceTimeout);
  },
  methods: {
    debounceSearch() {
      clearTimeout(this.debounceTimeout);

      // Immediate emit if input is cleared
      // Immediate search when clearing input
      if (!this.filters.searchText.trim()) {
        this.emitFilters();
        return;
      }

      // Faster debounce with progressive delay
      const baseDelay = 150;
      const textLength = this.filters.searchText.length;
      const dynamicDelay = textLength > 3 ? baseDelay : baseDelay * 2;

      this.debounceTimeout = setTimeout(() => {
        this.emitFilters();
      }, dynamicDelay);
    },
    emitFilters() {
      const payload = {
        searchText: this.filters.searchText.trim(),
        state: this.filters.state,
        charges: this.filters.charges,
        status: this.filters.status,
        affiliations: Object.keys(this.filters.affiliations).filter(
          (key) => this.filters.affiliations[key]
        ),
      };

      console.log("Emitting filters:", payload);
      this.$emit("filters-changed", payload);
    },
  },
};
</script>

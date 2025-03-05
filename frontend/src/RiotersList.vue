<!-- eslint-disable vue/html-self-closing -->

<template>
  <div class="flex-1 overflow-y-auto px-6">
    <!-- eslint-disable-next-line vue/max-attributes-per-line -->
    <ul v-if="riotersStore.filteredRioters.length > 0" class="space-y-4">
      <li
        v-for="rioter in riotersStore.filteredRioters"
        :key="rioter.id"
        :data-rioter-id="rioter.id"
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
          <button
            @click.stop="navigateToEdit(rioter)"
            class="px-3 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Edit
          </button>
        </div>
      </li>
    </ul>

    <!-- No Results Message -->

    <!-- eslint-disable-next-line vue/max-attributes-per-line -->
    <div v-else-if="!loading" class="bg-white shadow rounded-lg p-6 text-center mt-6">
      <p class="text-gray-500">No results found matching your filters.</p>
    </div>

    <!-- Loading State -->
    <!-- eslint-disable-next-line vue/max-attributes-per-line -->
    <div v-if="loading" class="text-center text-gray-500 py-6">Loading rioters...</div>
  </div>
</template>

<script setup>
defineProps({
  filteredRioters: {
    type: Array,
    default: () => [],
  },
  selectedRioter: {
    type: Object,
    default: () => null,
  },
  selectRioter: {
    type: Function,
    default: () => () => {},
  },
  getImageUrl: {
    type: Function,
    default: () => () => {},
  },
  handleImageError: {
    type: Function,
    default: () => () => {},
  },
  navigateToEdit: {
    type: Function,
    default: () => () => {},
  },
  loading: Boolean,
});
</script>

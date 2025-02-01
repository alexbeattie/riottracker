<template>
  <div class="p-6 flex items-center justify-center space-x-4 w-full">
    <!-- Previous Button -->
    <button
      @click="handlePageChange(currentPage - 1)"
      :disabled="currentPage === 1"
      class="px-4 py-2 bg-blue-500 text-white rounded-lg flex-1 disabled:bg-gray-300 disabled:cursor-not-allowed transition"
    >
      ← Prev
    </button>

    <!-- Page Indicator (Centered) -->
    <div class="flex-1 text-center text-gray-700 text-sm font-medium">
      Page <span class="font-semibold">{{ currentPage }}</span> of
      <span class="font-semibold">{{ totalPages }}</span>
    </div>

    <!-- Next Button -->
    <button
      @click="handlePageChange(currentPage + 1)"
      :disabled="currentPage === totalPages"
      class="px-2 py-2 bg-blue-500 text-white rounded-lg flex-1 disabled:bg-gray-300 disabled:cursor-not-allowed transition"
    >
      Next →
    </button>
  </div>
</template>

<script setup>
const props = defineProps({
  currentPage: { type: Number, default: 1 },
  totalPages: { type: Number, default: 10 },
  pageSize: { type: Number, default: 20 },
});

const emit = defineEmits(["page-changed"]);

const handlePageChange = (newPage) => {
  if (newPage >= 1 && newPage <= props.totalPages) {
    emit("page-changed", newPage);
  }
};
</script>

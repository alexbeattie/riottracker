<template>
  <div class="p-4">
    <div v-if="isLoading" class="text-center py-4">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"
      ></div>
      <p class="mt-2">Loading rioter data...</p>
    </div>
    <div v-else>
      <RioterForm mode="edit" :id="route.params.id" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import RioterForm from "../components/RioterForm.vue";
import { useRiotersStore } from "../stores/rioters"; // Note the plural
import { useRoute } from "vue-router";

const route = useRoute();
const riotersStore = useRiotersStore(); // Note the plural
const isLoading = ref(true);

onMounted(async () => {
  console.log("EditRioterView mounted with ID:", route.params.id);

  if (route.params.id) {
    try {
      isLoading.value = true;
      console.log("Fetching data for rioter:", route.params.id);
      await riotersStore.fetchRioterById(route.params.id); // Use the correct method
      console.log("Data fetched successfully:", riotersStore.selectedRioter);
    } catch (error) {
      console.error("Error fetching rioter data:", error);
    } finally {
      isLoading.value = false;
    }
  }
});
</script>

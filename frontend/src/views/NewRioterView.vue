<template>
  <div class="p-4">
    <RioterForm mode="new" />
  </div>
</template>

<script>
import RioterForm from "../components/RioterForm.vue";
import { useRiotersStore } from "../stores/rioters";

export default {
  name: "NewRioterView",
  components: {
    RioterForm,
  },
  setup() {
    // Initialize the store for this view
    const rioterStore = useRiotersStore();

    // Reset selectedRioter when creating a new rioter
    rioterStore.selectedRioter = null;

    return {
      rioterStore,
    };
  },
};
</script>

<style scoped>
/* Add view-specific styles if needed */
</style>

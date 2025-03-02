/* eslint-disable */
<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Add New Rioter</h1>
    <RioterForm mode="new" />
  </div>
</template>

<script>
import RioterForm from "../components/RioterForm.vue";

export default {
  name: "NewRioterView",
  components: {
    RioterForm,
  },
};
</script>

<style scoped>
/* Add view-specific styles if needed */
</style>

<template>
  <div class="p-4">
    <RioterForm mode="edit" :id="$route.params.id" />
  </div>
</template>

<script>
import RioterForm from "../components/RioterForm.vue";

export default {
  name: "EditRioterView",
  components: {
    RioterForm,
  },
};
</script>

<style scoped>
/* Add view-specific styles if needed */
</style>

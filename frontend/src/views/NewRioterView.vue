<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Add New Rioter</h1>
    <!-- Import your form component -->
    <NewRioterForm />
  </div>
</template>

<script>
import NewRioterForm from "../components/NewRioterForm.vue";

export default {
  name: "NewRioterView",
  components: {
    NewRioterForm,
  },
};
</script>

<style scoped>
/* Add view-specific styles if needed */
</style>

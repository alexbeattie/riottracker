<template>
  <nav class="bg-gray-800 p-4 flex justify-end">
    <ul class="flex space-x-4">
      <li>
        <router-link
          to="/"
          class="text-white hover:text-gray-300"
          active-class="font-bold"
        >
          Home
        </router-link>
      </li>
      <li>
        <router-link
          to="/map"
          class="text-white hover:text-gray-300"
          active-class="font-bold"
        >
          Map
        </router-link>
      </li>
      <li>
        <router-link
          to="/new"
          class="text-white hover:text-gray-300"
          active-class="font-bold"
        >
          Add New Rioter
        </router-link>
      </li>
    </ul>
  </nav>
</template>

<script>
export default {
  name: "Navigation",
};
</script>

<style scoped>
/* You can add extra styling here if needed */
</style>

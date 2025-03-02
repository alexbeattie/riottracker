<!-- src/components/RioterImage.vue -->
<template>
  <div class="relative">
    <img
      :src="imageUrl"
      :alt="alt"
      :class="imageClass"
      loading="lazy"
      @error="handleImageError"
    />
    <div
      v-if="isLoading"
      class="absolute inset-0 flex items-center justify-center bg-gray-100 rounded-full"
    >
      <div class="animate-spin rounded-full h-6 w-6 border-2 border-gray-500" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { getImageUrl, getPlaceholderUrl } from "../utils/imageHandling";

const props = defineProps({
  photoName: { type: String, default: "" },
  firstName: { type: String, default: "Unknown" },
  lastName: { type: String, default: "User" },

  imageClass: {
    type: String,
    default: "h-24 w-24 rounded-full object-cover border-2 border-gray-200",
  },
});

const isLoading = ref(true);
const hasError = ref(false);

const imageUrl = computed(() => {
  return typeof props.photoName === "string" && props.photoName.trim()
    ? getImageUrl(props.photoName.trim())
    : getPlaceholderUrl();
});
const alt = computed(() => `${props.firstName} ${props.lastName}`);

const handleImageError = (event) => {
  hasError.value = true;
  isLoading.value = false;
  event.target.src = getPlaceholderUrl();
};

// const onLoad = () => {
//   isLoading.value = false;
// };
</script>

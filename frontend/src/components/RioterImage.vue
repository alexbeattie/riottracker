<!-- src/components/RioterImage.vue -->
<template>
  <div class="relative">
    <img
      :src="imageUrl"
      :alt="alt"
      @error="handleImageError"
      :class="imageClass"
      loading="lazy"
    />
    <div
      v-if="isLoading"
      class="absolute inset-0 flex items-center justify-center bg-gray-100 rounded-full"
    >
      <div class="animate-spin rounded-full h-6 w-6 border-2 border-gray-500"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { getImageUrl, getPlaceholderUrl } from '../utils/imageHandling';

const props = defineProps({
  photoName: String,
  firstName: String,
  lastName: String,
  imageClass: {
    type: String,
    default: 'h-24 w-24 rounded-full object-cover border-2 border-gray-200'
  }
});

const isLoading = ref(true);
const hasError = ref(false);

const imageUrl = computed(() => getImageUrl(props.photoName));
const alt = computed(() => `${props.firstName} ${props.lastName}`);

const handleImageError = (event) => {
  hasError.value = true;
  isLoading.value = false;
  event.target.src = getPlaceholderUrl();
};

const onLoad = () => {
  isLoading.value = false;
};
</script>
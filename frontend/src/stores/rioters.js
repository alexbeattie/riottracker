import { defineStore } from "pinia";
import { ref, computed } from "vue";
import api from "../api"; // Ensure correct API path

export const useRiotersStore = defineStore("rioters", () => {
  const rioters = ref([]);
  const selectedRioter = ref(null);
  const searchText = ref("");
  const loading = ref(false);
  const error = ref(null);

  const fetchRioters = async () => {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.get("/rioters");
      rioters.value = response.data?.data || response.data;
    } catch (err) {
      error.value = err.message || "Failed to fetch rioters";
    } finally {
      loading.value = false;
    }
  };

  const fetchRioterById = async (id) => {
    loading.value = true;
    error.value = null;
    try {
      console.log(`Fetching rioter with ID: ${id}`);
      const response = await api.get(`/rioters/${id}`);
      console.log("API response:", response);
      selectedRioter.value = response.data;
      console.log("Selected rioter set to:", selectedRioter.value);
      return response.data;
    } catch (err) {
      console.error(`Failed to fetch rioter with ID ${id}:`, err);
      error.value = err.message || `Failed to fetch rioter with ID ${id}`;
      throw err;
    } finally {
      loading.value = false;
    }
  };
  const filteredRioters = computed(() => {
    return rioters.value.filter((rioter) =>
      searchText.value
        ? `${rioter.first_name} ${rioter.last_name}`
          .toLowerCase()
          .includes(searchText.value.toLowerCase())
        : true
    );
  });

  return {
    rioters,
    selectedRioter,
    searchText,
    loading,
    error,
    fetchRioters,
    fetchRioterById,
    filteredRioters,
  };
});
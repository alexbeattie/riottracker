import { defineStore } from "pinia";
import axios from "axios";

const API_BASE_URL = "http://localhost:8080";

export const useRioterStore = defineStore("rioters", {
  state: () => ({
    rioters: [], // Stores the list of rioters
    selectedRioter: null, // Stores the currently edited rioter
  }),

  actions: {
    async fetchRioters() {
      try {
        const { data } = await axios.get(`${API_BASE_URL}/api/rioters`);
        this.rioters = data.data;
      } catch (error) {
        console.error("Error fetching rioters:", error);
      }
    },

    async fetchRioter(id) {
      try {
        const { data } = await axios.get(`${API_BASE_URL}/api/rioters/${id}`);
        this.selectedRioter = data;
      } catch (error) {
        console.error("Error fetching rioter:", error);
      }
    },

    async updateRioter(updatedRioter) {
      try {
        await axios.put(`${API_BASE_URL}/api/rioters/${updatedRioter.id}`, updatedRioter);

        // ✅ Update local state immediately (avoids full page reload)
        const index = this.rioters.findIndex((r) => r.id === updatedRioter.id);
        if (index !== -1) {
          this.rioters[index] = { ...updatedRioter };
        }

        this.selectedRioter = { ...updatedRioter };

        console.log("✅ Rioter updated in state:", updatedRioter);
      } catch (error) {
        console.error("Error updating rioter:", error);
      }
    },
  },
});

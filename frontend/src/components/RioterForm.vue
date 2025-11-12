<!-- eslint-disable -->
<template>
  <div class="max-w-lg mx-auto p-4 border rounded shadow">
    <!-- When in edit mode, show a loading indicator while fetching data -->
    <div v-if="mode === 'edit' && loading" class="text-center py-4">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"
      ></div>
      <p class="mt-2">Loading rioter data...</p>
    </div>

    <!-- Form content -->
    <template v-else>
      <h2 class="text-2xl font-bold mb-4">
        {{ mode === "edit" ? "Update Rioter" : "Add New Rioter" }}
      </h2>
      <form @submit.prevent="submitForm">
        <!-- First Name -->
        <div class="mb-4">
          <label for="firstName" class="block font-medium">First Name</label>
          <input
            id="firstName"
            type="text"
            v-model="form.first_name"
            class="w-full p-2 border rounded"
            required
          />
        </div>

        <!-- Last Name -->
        <div class="mb-4">
          <label for="lastName" class="block font-medium">Last Name</label>
          <input
            id="lastName"
            type="text"
            v-model="form.last_name"
            class="w-full p-2 border rounded"
            required
          />
        </div>

        <!-- Middle Name -->
        <div class="mb-4">
          <label for="middleName" class="block font-medium">Middle Name</label>
          <input
            id="middleName"
            type="text"
            v-model="form.middle_name"
            class="w-full p-2 border rounded"
          />
        </div>

        <!-- Photo Upload -->
        <div class="mb-4">
          <label for="photo" class="block font-medium">Profile Photo</label>
          <div class="flex items-center space-x-4">
            <img
              :src="photoPreview || getImageUrl(form.photo_name)"
              class="h-32 w-32 rounded-full object-cover border border-gray-200"
              @error="handleImageError"
            />
            <div>
              <input
                id="photo"
                type="file"
                accept="image/jpeg,image/png,image/jpg"
                @change="handlePhotoChange"
                class="hidden"
                ref="photoInput"
              />
              <button
                type="button"
                @click="$refs.photoInput.click()"
                class="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
              >
                Change Photo
              </button>
              <p v-if="form.photo_name" class="text-sm mt-1 text-gray-500">
                Current: {{ form.photo_name }}
              </p>
            </div>
          </div>
        </div>

        <!-- Summary -->
        <div class="mb-4">
          <label for="summary" class="block font-medium">Summary</label>
          <textarea
            id="summary"
            v-model="form.summary"
            class="w-full p-2 border rounded"
          ></textarea>
        </div>

        <!-- Jurisdiction -->
        <div class="mb-4">
          <label for="jurisdiction" class="block font-medium">Jurisdiction</label>
          <input
            id="jurisdiction"
            type="text"
            v-model="form.jurisdiction"
            class="w-full p-2 border rounded"
          />
        </div>

        <!-- Charges -->
        <div class="mb-4">
          <label for="charges" class="block font-medium">Charges</label>
          <textarea
            id="charges"
            v-model="form.charges"
            class="w-full p-2 border rounded"
          ></textarea>
        </div>

        <!-- Charges Link -->
        <div class="mb-4">
          <label for="charges_link" class="block font-medium">Charges Link</label>
          <input
            id="charges_link"
            type="url"
            v-model="form.charges_link"
            class="w-full p-2 border rounded"
          />
        </div>

        <!-- Case Status -->
        <div class="mb-4">
          <label for="case_status" class="block font-medium">Case Status</label>
          <input
            id="case_status"
            type="text"
            v-model="form.case_status"
            class="w-full p-2 border rounded"
          />
        </div>

        <!-- Case Updates -->
        <div class="mb-4">
          <label for="case_updates" class="block font-medium">Case Updates</label>
          <textarea
            id="case_updates"
            v-model="form.case_updates"
            class="w-full p-2 border rounded"
          ></textarea>
        </div>

        <!-- Boolean Checkboxes -->
        <div class="mb-4" v-for="(label, key) in booleanFields" :key="key">
          <label class="inline-flex items-center">
            <input type="checkbox" v-model="form[key]" class="mr-2" />
            {{ label }}
          </label>
        </div>

        <!-- Age -->
        <div class="mb-4">
          <label for="age" class="block font-medium">Age</label>
          <input
            id="age"
            type="number"
            v-model.number="form.age"
            class="w-full p-2 border rounded"
          />
        </div>

        <!-- City -->
        <div class="mb-4">
          <label for="city" class="block font-medium">City</label>
          <input
            id="city"
            type="text"
            v-model="form.city"
            class="w-full p-2 border rounded"
          />
        </div>

        <!-- State -->
        <div class="mb-4">
          <label for="state" class="block font-medium">State</label>
          <input
            id="state"
            type="text"
            v-model="form.state"
            class="w-full p-2 border rounded"
          />
        </div>

        <!-- Submit and Cancel Buttons -->
        <div class="flex gap-4">
          <button
            type="submit"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            :disabled="submitting"
          >
            <span v-if="submitting" class="inline-block animate-spin mr-2">⟳</span>
            {{ mode === "edit" ? "Update" : "Submit" }}
          </button>
          <!-- Only show Cancel if in edit mode -->
          <button
            v-if="mode === 'edit'"
            type="button"
            @click="cancel"
            class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
          >
            Cancel
          </button>
        </div>
      </form>

      <!-- Success and Error Messages -->
      <div v-if="message" class="mt-4 p-2 bg-green-100 text-green-700 rounded">
        {{ message }}
      </div>
      <div v-if="error" class="mt-4 p-2 bg-red-100 text-red-700 rounded">
        {{ error }}
      </div>
    </template>
  </div>
</template>

<script>
import { useRiotersStore } from "@/stores/rioters";
import { storeToRefs } from "pinia";
import api from "@/api"; // Make sure to import the API
// import axios from "axios";

export default {
  name: "RioterForm",
  props: {
    // mode can be 'new' or 'edit'
    mode: {
      type: String,
      default: "new",
    },
    // For edit mode, pass the rioter id
    id: {
      type: String,
      default: null,
    },
  },
  data() {
    return {
      form: {
        first_name: "",
        last_name: "",
        middle_name: "",
        summary: "",
        jurisdiction: "",
        charges: "",
        charges_link: "",
        case_status: "",
        case_updates: "",
        violence_assault: false,
        conspiracy: false,
        theft: false,
        property: false,
        age: null,
        city: "",
        state: "",
        military_le: false,
        extremist: false,
        sentenced: false,
        inspired_trump: false,
        commuted: false,
        pardoned: false,
        arrest_date: "",
        photo_name: "", // Include the current photo name if available
        // Include id if we're in edit mode
        id: this.mode === "edit" ? this.id : null,
      },
      photoFile: null, // Add this for the photo upload
      photoPreview: null, // Add this for the photo preview
      message: "",
      error: "",
      loading: false,
      submitting: false,
      // A mapping for the boolean fields to show custom labels
      booleanFields: {
        violence_assault: "Violence/Assault",
        conspiracy: "Conspiracy",
        theft: "Theft",
        property: "Property Damage",
        military_le: "Military/LE",
        extremist: "Extremist",
        inspired_trump: "Inspired Trump",
        sentenced: "Sentenced",
        commuted: "Commuted",
        pardoned: "Pardoned",
      },
    };
  },
  setup() {
    const riotersStore = useRiotersStore();
    const { selectedRioter } = storeToRefs(riotersStore);

    return {
      riotersStore,
      selectedRioter,
    };
  },
  watch: {
    // Watch for changes in the store's selectedRioter
    selectedRioter(newValue) {
      if (newValue && this.mode === "edit") {
        this.updateFormFromSelectedRioter();
      }
    },
  },
  mounted() {
    console.log("RioterForm mounted - mode:", this.mode, "id:", this.id);

    if (this.mode === "edit" && this.id) {
      this.loadRioter();
    }
  },
  methods: {
    async loadRioter() {
      console.log("loadRioter started");

      this.loading = true;
      try {
        console.log("Fetching rioter with ID:", this.id);

        // Use the correct method name from your store
        await this.riotersStore.fetchRioterById(this.id);
        console.log("Fetch complete, selectedRioter:", this.selectedRioter);
        if (this.selectedRioter) {
          this.updateFormFromSelectedRioter();
        } else {
          console.warn("selectedRioter is null after fetch");
        }
      } catch (error) {
        this.error = "Failed to load rioter data";
        console.error("Error loading rioter:", error);
      } finally {
        this.loading = false;
        console.log("loadRioter complete, loading state:", this.loading);
      }
    },
    updateFormFromSelectedRioter() {
      if (this.selectedRioter) {
        const photoName = this.selectedRioter.photo_name;

        Object.keys(this.form).forEach((key) => {
          if (this.selectedRioter[key] !== undefined) {
            this.form[key] = this.selectedRioter[key];
          }
        });
        if (photoName) {
          this.form.photo_name = photoName;
          console.log("Photo name set:", this.form.photo_name);
        }
        // Ensure the ID is set
        this.form.id = this.selectedRioter.id;
      }
    },
    handlePhotoChange(event) {
      const file = event.target.files[0];
      if (!file) return;

      this.photoFile = file;

      // Create a preview
      const reader = new FileReader();
      reader.onload = (e) => {
        this.photoPreview = e.target.result;
      };
      reader.readAsDataURL(file);
    },
    async uploadPhoto() {
      if (!this.photoFile) return null;

      // Create form data
      const formData = new FormData();
      formData.append("photo", this.photoFile);
      formData.append("id", this.form.id);

      try {
        // Use the correct endpoint with /api prefix
        const response = await fetch("http://localhost:8085/api/rioters/upload-photo", {
          method: "POST",
          body: formData,
        });

        if (!response.ok) {
          throw new Error(`Upload failed: ${response.status}`);
        }

        const data = await response.json();
        return data.photo_name;
      } catch (error) {
        console.error("Photo upload error:", error);
        throw error;
      }
    },
    getImageUrl(photoName) {
      if (!photoName) {
        // Return inline SVG as fallback
        return "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100' viewBox='0 0 100 100'%3E%3Crect width='100' height='100' fill='%23cccccc'/%3E%3Ctext x='50' y='50' font-size='14' text-anchor='middle' alignment-baseline='middle' font-family='Arial' fill='%23666666'%3ENo Image%3C/text%3E%3C/svg%3E";
      }

      // Fix: Use the correct URL for photos (correcting the port to 8085)
      return `http://localhost:8085/photos/${encodeURIComponent(photoName)}`;
    },

    handleImageError(event) {
      event.target.src =
        "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100' viewBox='0 0 100 100'%3E%3Crect width='100' height='100' fill='%23cccccc'/%3E%3Ctext x='50' y='50' font-size='14' text-anchor='middle' alignment-baseline='middle' font-family='Arial' fill='%23666666'%3ENo Image%3C/text%3E%3C/svg%3E";
    },
    async submitForm() {
      this.message = "";
      this.error = "";
      this.submitting = true;

      try {
        // Handle photo upload if a new photo was selected
        if (this.photoFile) {
          const newPhotoName = await this.uploadPhoto();
          if (newPhotoName) {
            this.form.photo_name = newPhotoName;
          }
        }

        if (this.mode === "edit") {
          // Make sure photo_name is included in the update
          console.log("Updating with photo_name:", this.form.photo_name);

          // Use direct API call instead of store method
          await api.put(`/rioters/${this.form.id}`, this.form);
          this.message = "Rioter record updated successfully!";

          // Optionally redirect after update
          setTimeout(() => {
            this.$router.push("/");
          }, 1500);
        } else {
          // Create new rioter
          await api.post("/rioters", this.form);
          this.message = "Rioter record added successfully!";
          this.resetForm();
        }
      } catch (error) {
        this.error = error.response?.data?.error || error.message || "Submission failed";
        console.error("Form submission error:", error);
      } finally {
        this.submitting = false;
      }
    },
    resetForm() {
      // Reset the form to initial values
      this.form = {
        first_name: "",
        last_name: "",
        middle_name: "",
        summary: "",
        jurisdiction: "",
        charges: "",
        charges_link: "",
        case_status: "",
        case_updates: "",
        violence_assault: false,
        conspiracy: false,
        theft: false,
        property: false,
        age: null,
        city: "",
        state: "",
        military_le: false,
        extremist: false,
        sentenced: false,
        inspired_trump: false,
        commuted: false,
        pardoned: false,
        arrest_date: "",
      };
      this.photoFile = null;
      this.photoPreview = null;
    },
    cancel() {
      // In edit mode, cancel and navigate away
      this.$router.push("/");
    },
  },
};
</script>

<style scoped>
/* Add any additional styling if needed */
</style>

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
          >
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
import api from "@/api";

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
      // All the fields needed for a rioter
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
      },
      message: "",
      error: "",
      loading: false,
      // A mapping for the boolean fields to show custom labels
      booleanFields: {
        violence_assault: "Violence/Assault",
        conspiracy: "Conspiracy",
        theft: "Theft",
        property: "Property Damage",
        military_le: "Military/LE",
        extremist: "Extremist",
        inspired_trump: "Inspired Trump",
      },
    };
  },
  mounted() {
    if (this.mode === "edit" && this.id) {
      this.loadRioter();
    }
  },
  methods: {
    async loadRioter() {
      this.loading = true;
      try {
        const response = await api.get(`/rioters/${this.id}`);
        // Update form with fetched data
        Object.keys(this.form).forEach((key) => {
          if (response.data[key] !== undefined) {
            this.form[key] = response.data[key];
          }
        });
      } catch (error) {
        this.error = "Failed to load rioter data";
        console.error("Error loading rioter:", error);
      } finally {
        this.loading = false;
      }
    },
    async submitForm() {
      this.message = "";
      this.error = "";
      try {
        if (this.mode === "edit") {
          await api.put(`/rioters/${this.id}`, this.form);
          this.message = "Rioter record updated successfully!";
          // Optionally redirect after update
          setTimeout(() => {
            this.$router.push("/");
          }, 1500);
        } else {
          await api.post("/rioters", this.form);
          this.message = "Rioter record added successfully!";
          this.resetForm();
        }
      } catch (error) {
        this.error = error.response?.data?.error || error.message || "Submission failed";
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

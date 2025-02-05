<template>
  <div class="max-w-lg mx-auto p-4 border rounded shadow">
    <h2 class="text-2xl font-bold mb-4">Edit Rioter</h2>
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
      <div class="mb-4">
        <label class="inline-flex items-center">
          <input type="checkbox" v-model="form.violence_assault" class="mr-2" />
          Violence/Assault
        </label>
      </div>
      <div class="mb-4">
        <label class="inline-flex items-center">
          <input type="checkbox" v-model="form.conspiracy" class="mr-2" />
          Conspiracy
        </label>
      </div>
      <div class="mb-4">
        <label class="inline-flex items-center">
          <input type="checkbox" v-model="form.theft" class="mr-2" />
          Theft
        </label>
      </div>
      <div class="mb-4">
        <label class="inline-flex items-center">
          <input type="checkbox" v-model="form.property" class="mr-2" />
          Property Damage
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

      <!-- Additional Boolean Fields -->
      <div class="mb-4">
        <label class="inline-flex items-center">
          <input type="checkbox" v-model="form.military_le" class="mr-2" />
          Military/LE
        </label>
      </div>
      <div class="mb-4">
        <label class="inline-flex items-center">
          <input type="checkbox" v-model="form.extremist" class="mr-2" />
          Extremist
        </label>
      </div>
      <div class="mb-4">
        <label class="inline-flex items-center">
          <input type="checkbox" v-model="form.inspired_trump" class="mr-2" />
          Inspired Trump
        </label>
      </div>

      <!-- Hidden field for photo_name so it is always sent -->
      <input type="hidden" v-model="form.photo_name" />

      <!-- Submit Button -->
      <button type="submit" class="px-4 py-2 bg-blue-600 text-white rounded">
        Update
      </button>
    </form>

    <!-- Success and Error Messages -->
    <div v-if="message" class="mt-4 text-green-600">{{ message }}</div>
    <div v-if="error" class="mt-4 text-red-600">{{ error }}</div>
  </div>
</template>

<script>
import api from "@/api"; // adjust the path if necessary

export default {
  name: "EditRioterForm",
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
        // Include the current photo name if available.
        photo_name: "",
      },
      message: "",
      error: "",
    };
  },
  mounted() {
    // Retrieve the current record and merge photo_name into the form.
    this.fetchRioterData();
  },
  methods: {
    async fetchRioterData() {
      try {
        const response = await api.get(`/rioters/${this.$route.params.id}`);
        // Merge the fetched data into the form.
        Object.assign(this.form, response.data);
      } catch (error) {
        this.error = "Failed to load rioter data";
        console.error("Error fetching rioter:", error);
      }
    },
    async submitForm() {
      this.message = "";
      this.error = "";
      try {
        // Update the rioter record. The payload now always includes photo_name.
        await api.put(`/rioters/${this.$route.params.id}`, this.form);
        this.message = "Rioter record updated successfully!";
      } catch (error) {
        this.error =
          (error.response && error.response.data && error.response.data.error) ||
          error.message;
      }
    },
  },
};
</script>

<style scoped>
/* Add any additional styling here */
</style>

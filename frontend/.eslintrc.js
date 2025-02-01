module.exports = {
  env: {
    node: true,
  },
  extends: [
    "plugin:vue/vue3-recommended", // Ensure this is included
    "eslint:recommended",
  ],
  parserOptions: {
    ecmaVersion: 2020,
    sourceType: "module",
  },
  rules: {
    "vue/multi-word-component-names": "off", // Disable multi-word rule if needed
  },
  globals: {
    defineProps: "readonly", // Add this
    defineEmits: "readonly", // Add this
    defineExpose: "readonly", // Add this (optional, for future use)
  },
};
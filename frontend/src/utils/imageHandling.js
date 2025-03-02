const API_BASE_URL = "http://localhost:8080";

export const getImageUrl = (photoName) => {
  if (!photoName || typeof photoName !== "string") {
    return `${API_BASE_URL}/photos/placeholder.jpg`; // ✅ Always return a valid string
  }
  return `${API_BASE_URL}/photos/${encodeURIComponent(photoName.trim())}`;
};

export const getPlaceholderUrl = () => {
  return `${API_BASE_URL}/photos/placeholder.jpg`;
};

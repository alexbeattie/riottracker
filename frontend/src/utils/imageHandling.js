const API_BASE_URL = 'http://localhost:8080';

export const getImageUrl = (photoName) => {
  if (!photoName?.trim()) {
    return `${API_BASE_URL}/photos/placeholder.jpg`;
  }
  return `${API_BASE_URL}/photos/${encodeURIComponent(photoName)}`;
};

export const getPlaceholderUrl = () => {
  return `${API_BASE_URL}/photos/placeholder.jpg`;
};


const API_BASE_URL = import.meta.env.VITE_API_URL || '';
const UPLOADS_BASE_URL = import.meta.env.VITE_UPLOADS_URL || '';

export const API_ENDPOINTS = {
    LOGIN: `${API_BASE_URL}/api/login`,
    REGISTER: `${API_BASE_URL}/api/register`,
    MOVIES: `${API_BASE_URL}/api/movies`,
    STREAM: `${API_BASE_URL}/api/stream`,
    THUMBNAIL: `${API_BASE_URL}/api/movies`,
    UPLOADS: `${API_BASE_URL}/api/movies`,
    ADMIN: `${API_BASE_URL}/api/admin`,
    ME: `${API_BASE_URL}/api/me`,
    BROWSE: `${API_BASE_URL}/api/browse`,
    WATCH_HISTORY: `${API_BASE_URL}/api/watch-history`,
    RECOMMENDATIONS: `${API_BASE_URL}/api/recommendations`
};

export default { API_BASE_URL, UPLOADS_BASE_URL };

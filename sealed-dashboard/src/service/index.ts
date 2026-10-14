import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";

const baseURL = import.meta.env.VITE_SERVER_URL;

const api = axios.create({
  baseURL,
  timeout: 30 * 1000,
  headers: { "Content-Type": "application/json" },
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("access_token");
  if (token) config.headers.set("Authorization", `Bearer ${token}`);
  return config;
});

let refreshPromise: Promise<string> | null = null;

function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    const refreshToken = localStorage.getItem("refresh_token");
    if (!refreshToken) {
      return Promise.reject(new Error("No refresh token"));
    }

    refreshPromise = axios
      .post(`${baseURL}/auth/revalidate-token`, null, {
        headers: { Authorization: `Bearer ${refreshToken}` },
      })
      .then((res) => {
        const { access_token, refresh_token } = res.data.data; // adjust if wrapped in .data
        localStorage.setItem("access_token", access_token);
        localStorage.setItem("refresh_token", refresh_token);
        return access_token as string;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

type RetryableConfig = InternalAxiosRequestConfig & { _retry?: boolean };

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const original = error.config as RetryableConfig | undefined;
    const isRefreshCall = original?.url?.includes("/auth/revalidate-token");

    if (error.response?.status !== 401 || !original || original._retry || isRefreshCall) {
      return Promise.reject(error);
    }

    original._retry = true;

    try {
      const newAccessToken = await refreshAccessToken();
      return api.request({
        ...original,
        headers: { ...original.headers, Authorization: `Bearer ${newAccessToken}` },
      });
    } catch (refreshError) {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
      window.dispatchEvent(new Event("unauthorized"));
      return Promise.reject(refreshError);
    }
  },
);

export default api;

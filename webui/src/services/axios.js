import axios from "axios";

// NOTE: Usually this is injected by Vite as __API_URL__, but for the homework:
const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});

// on startup, load token if exists
const storedToken = localStorage.getItem("token");
if (storedToken) {
	instance.defaults.headers.common["Authorization"] = `Bearer ${storedToken}`;
}

export function setToken(token) {
	localStorage.setItem("token", token);
	instance.defaults.headers.common["Authorization"] = `Bearer ${token}`;
}

export function clearToken() {
	localStorage.removeItem("token");
	delete instance.defaults.headers.common["Authorization"];
}

export default instance;

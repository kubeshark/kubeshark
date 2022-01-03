import * as axios from "axios";

// When working locally cp `cp .env.example .env`
export const MizuWebsocketURL = process.env.REACT_APP_OVERRIDE_WS_URL ? process.env.REACT_APP_OVERRIDE_WS_URL :
                        window.location.protocol === 'https:' ? `wss://${window.location.host}/ws` : `ws://${window.location.host}/ws`;

const CancelToken = axios.CancelToken;

// When working locally cp `cp .env.example .env`
const apiURL = process.env.REACT_APP_OVERRIDE_API_URL ? process.env.REACT_APP_OVERRIDE_API_URL : `${window.location.origin}/`;

export default class Api {

    constructor() {
        this.token = localStorage.getItem("token");

        this.client = this.getAxiosClient();
        this.source = null;
    }

    tapStatus = async () => {
        const response = await this.client.get("/status/tap");
        return response.data;
    }

    analyzeStatus = async () => {
        const response = await this.client.get("/status/analyze");
        return response.data;
    }

    getEntry = async (id) => {
        const response = await this.client.get(`/entries/${id}`);
        return response.data;
    }

    fetchEntries = async (leftOff, direction, query, limit, timeoutMs) => {
        const response = await this.client.get(`/entries/?leftOff=${leftOff}&direction=${direction}&query=${query}&limit=${limit}&timeoutMs=${timeoutMs}`).catch(function (thrown) {
            console.error(thrown.message);
            return {};
        });
        return response.data;
    }

    getRecentTLSLinks = async () => {
        const response = await this.client.get("/status/recentTLSLinks");
        return response.data;
    }

    getAuthStatus = async () => {
        const response = await this.client.get("/status/auth");
        return response.data;
    }

    validateQuery = async (query) => {
        if (this.source) {
            this.source.cancel();
        }
        this.source = CancelToken.source();

        const form = new FormData();
        form.append('query', query)
        const response = await this.client.post(`/query/validate`, form, {
            cancelToken: this.source.token
        }).catch(function (thrown) {
            if (!axios.isCancel(thrown)) {
                console.error('Validate error', thrown.message);
            }
        });

        if (!response) {
            return null;
        }

        return response.data;
    }

    isInstallNeeded = async () => {
        const response = await this.client.get("/install/isNeeded");
        return response.data;
    }

    isAuthenticationNeeded = async () => {
        try {
            await this.client.get("/status/tap");
            return false;
        } catch (e) {
            if (e.response.status == 401) {
                return true;
            }
            throw e;
        }
    }

    postInstall = async (adminPassword) => {
        const form = new FormData();
        form.append('adminPassword', adminPassword)

        const response = await this.client.post(`/install/`, form);
        if (response.status >= 200 && response.status < 300) {
            this.persistToken(response.data.token);
        }

    }

    login = async (username, password) => {
        const form = new FormData();
        form.append('username', username);
        form.append('password', password);

        const response = await this.client.post(`/user/login`, form);
        if (response.status >= 200 && response.status < 300) {
            this.persistToken(response.data.token);
        }

        return response;
    }

    persistToken = (token) => {
        this.token = token;
        this.client = this.getAxiosClient();
        localStorage.setItem('token', token);
    }

    logout = () => {
        this.persistToken(null);
    }

    getAxiosClient = () => {
        const headers = {
            Accept: "application/json"
        }

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        return axios.create({
            baseURL: apiURL,
            timeout: 31000,
            headers
        });
    }
}

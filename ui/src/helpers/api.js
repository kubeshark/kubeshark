import * as axios from "axios";

export const MizuWebsocketURL = process.env.REACT_APP_OVERRIDE_WS_URL ? process.env.REACT_APP_OVERRIDE_WS_URL :
    window.location.protocol === 'https:' ? `wss://${window.location.host}/ws` : `ws://${window.location.host}/ws`;

export const FormValidationErrorType = "formError";

const CancelToken = axios.CancelToken;

const apiURL = process.env.REACT_APP_OVERRIDE_API_URL ? process.env.REACT_APP_OVERRIDE_API_URL : `${window.location.origin}/`;

export default class Api {
    static instance;

    static getInstance() {
        if (!Api.instance) {
            Api.instance = new Api();
        }
        return Api.instance;
    }

    constructor() {
        this.token = localStorage.getItem("token");

        this.client = this.getAxiosClient();
        this.source = null;
    }

    serviceMapStatus = async () => {
        const response = await this.client.get("/servicemap/status");
        return response.data;
    }

    serviceMapData = async () => {
        const response = await this.client.get(`/servicemap/get`);
        return response.data;
    }

    serviceMapReset = async () => {
        const response = await this.client.get(`/servicemap/reset`);
        return response.data;
    }

    tapStatus = async () => {
        const response = await this.client.get("/status/tap");
        return response.data;
    }

    analyzeStatus = async () => {
        const response = await this.client.get("/status/analyze");
        return response.data;
    }

    getEntry = async (id, query) => {
        const response = await this.client.get(`/entries/${id}?query=${query}`);
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

    getOasServices = async () => {
        const response = await this.client.get("/oas");
        return response.data;
    }

    getOasByService = async (selectedService) => {
        const response = await this.client.get(`/oas/${selectedService}`);
        return response.data;
    }

    gelAlloasServicesInOneSpec = async () => {
        const response = await this.client.get("/oas/all");
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

    getTapConfig = async () => {
        const response = await this.client.get("/config/tap");
        return response.data;
    }

    setTapConfig = async (config) => {
        const response = await this.client.post("/config/tap", {tappedNamespaces: config});
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
            if (e.response.status === 401) {
                return true;
            }
            throw e;
        }
    }

    setupAdminUser = async (password) => {
        const form = new FormData();
        form.append('password', password);

        try {
            const response = await this.client.post(`/install/admin`, form);
            this.persistToken(response.data.token);
            return response;
        } catch (e) {
            if (e.response.status === 400) {
                const error = {
                    'type': FormValidationErrorType,
                    'messages': e.response.data
                };
                throw error;
            } else {
                throw e;
            }
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

    logout = async () => {
        await this.client.post(`/user/logout`);
        this.persistToken(null);
    }

    getAxiosClient = () => {
        const headers = {
            Accept: "application/json"
        }

        if (this.token) {
            headers['x-session-token'] = `${this.token}`; // we use `x-session-token` instead of `Authorization` because the latter is reserved by kubectl proxy, making mizu view not work
        }
        return axios.create({
            baseURL: apiURL,
            timeout: 31000,
            headers
        });
    }
}

import * as axios from "axios";

// When working locally cp `cp .env.example .env`
export const MizuWebsocketURL = process.env.REACT_APP_OVERRIDE_WS_URL ? process.env.REACT_APP_OVERRIDE_WS_URL :
                        window.location.protocol === 'https:' ? `wss://${window.location.host}/ws` : `ws://${window.location.host}/ws`;

// When working locally cp `cp .env.example .env`
export const MizuApiURL = process.env.REACT_APP_OVERRIDE_API_URL ? process.env.REACT_APP_OVERRIDE_API_URL : `${window.location.origin}/`;

const CancelToken = axios.CancelToken;

export default class Api {

    constructor() {

        this.client = axios.create({
            baseURL: MizuApiURL,
            timeout: 31000,
            headers: {
                Accept: "application/json",
            }
        });

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
}

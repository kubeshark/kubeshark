import * as axios from "axios";

export const MizuWebsocketURL = process.env.REACT_APP_OVERRIDE_WS_URL ? process.env.REACT_APP_OVERRIDE_WS_URL :
    window.location.protocol === 'https:' ? `wss://${window.location.host}/ws` : `ws://${window.location.host}/ws`;

export const FormValidationErrorType = "formError";

const CancelToken = axios.CancelToken;

const apiURL = process.env.REACT_APP_OVERRIDE_API_URL ? process.env.REACT_APP_OVERRIDE_API_URL : `${window.location.origin}/`;

let token = ""
let client = null
let source = null

export default class Api {
    static instance;

    static getInstance() {
        if (!Api.instance) {
            Api.instance = new Api();
        }
        return Api.instance;
    }

    constructor() {
        token = localStorage.getItem("token");

        client = this.getAxiosClient();
        source = null;
    }

    serviceMapStatus = async () => {
        const response = await client.get("/servicemap/status");
        return response.data;
    }

    serviceMapData = async () => {
        const response = await client.get(`/servicemap/get`);
        return response.data;
    }

    serviceMapReset = async () => {
        const response = await client.get(`/servicemap/reset`);
        return response.data;
    }

    tapStatus = async () => {
        const response = await client.get("/status/tap");
        return response.data;
    }
    getTapConfig = async () => {
        const response = await this.client.get("/config/tap");
        return response.data;
    }

    setTapConfig = async (config) => {
        const response = await this.client.post("/config/tap", { tappedNamespaces: config });
        return response.data;
    }

    analyzeStatus = async () => {
        const response = await client.get("/status/analyze");
        return response.data;
    }

    getEntry = async (id, query) => {
        const response = await client.get(`/entries/${id}?query=${query}`);
        return response.data;
    }

    fetchEntries = async (leftOff, direction, query, limit, timeoutMs) => {
        const response = await client.get(`/entries/?leftOff=${leftOff}&direction=${direction}&query=${query}&limit=${limit}&timeoutMs=${timeoutMs}`).catch(function (thrown) {
            console.error(thrown.message);
            return {};
        });
        return response.data;
    }

    getRecentTLSLinks = async () => {
        const response = await client.get("/status/recentTLSLinks");
        return response.data;
    }

    getAuthStatus = async () => {
        const response = await client.get("/status/auth");
        return response.data;
    }

    getOasServices = async () => {
        const response = await client.get("/oas/");
        return response.data;
    }

    getOasByService = async (selectedService) => {
        const response = await client.get(`/oas/${selectedService}`);
        return response.data;
    }

    gelAlloasServicesInOneSpec = async () => {
        const response = await this.client.get("/oas/all");
        return response.data;
    }

    validateQuery = async (query) => {
        if (source) {
            source.cancel();
        }
        source = CancelToken.source();

        const form = new FormData();
        form.append('query', query)
        const response = await client.post(`/query/validate`, form, {
            cancelToken: source.token
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

    persistToken = (tk) => {
        token = tk;
        client = this.getAxiosClient();
        localStorage.setItem('token', token);
    }

    getAxiosClient = () => {
        const headers = {
            Accept: "application/json"
        }

        if (token) {
            headers['x-session-token'] = `${token}`; // we use `x-session-token` instead of `Authorization` because the latter is reserved by kubectl proxy, making mizu view not work
        }
        return axios.create({
            baseURL: apiURL,
            timeout: 31000,
            headers
        });
    }
}

export function getWebsocketUrl() {
    let websocketUrl = MizuWebsocketURL;
    if (token) {
        websocketUrl += `/${token}`;
    }

    return websocketUrl;
}
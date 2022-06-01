import * as axios from "axios";

export const MizuWebsocketURL = process.env.REACT_APP_OVERRIDE_WS_URL ? process.env.REACT_APP_OVERRIDE_WS_URL :
    window.location.protocol === 'https:' ? `wss://${window.location.host}/ws` : `ws://${window.location.host}/ws`;

const CancelToken = axios.CancelToken;

const apiURL = process.env.REACT_APP_OVERRIDE_API_URL ? process.env.REACT_APP_OVERRIDE_API_URL : `${window.location.origin}/`;

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
        client = this.getAxiosClient();
        source = null;
    }

    serviceMapData = async () => {
        const response = await client.get(`/servicemap/get`);
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

    getEntry = async (id, query) => {
        const response = await client.get(`/entries/${id}?query=${encodeURIComponent(query)}`);
        return response.data;
    }

    fetchEntries = async (leftOff, direction, query, limit, timeoutMs) => {
        const response = await client.get(`/entries/?leftOff=${leftOff}&direction=${direction}&query=${encodeURIComponent(query)}&limit=${limit}&timeoutMs=${timeoutMs}`).catch(function (thrown) {
            console.error(thrown.message);
            return {};
        });
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

    getAxiosClient = () => {
        const headers = {
            Accept: "application/json"
        }
        return axios.create({
            baseURL: apiURL,
            timeout: 31000,
            headers
        });
    }
}

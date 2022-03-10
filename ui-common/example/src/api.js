import * as axios from "axios";

export const MizuWebsocketURL = process.env.REACT_APP_OVERRIDE_WS_URL ? process.env.REACT_APP_OVERRIDE_WS_URL :
    window.location.protocol === 'https:' ? `wss://${window.location.host}/ws` : `ws://${window.location.host}/ws`;

export const FormValidationErrorType = "formError";

const CancelToken = axios.CancelToken;

const apiURL = process.env.REACT_APP_OVERRIDE_API_URL ? process.env.REACT_APP_OVERRIDE_API_URL : `${window.location.origin}/`;

let token = null
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

    tapStatus = async () => {
        const response = await client.get("/status/tap");
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
        const response = await client.get("/oas");
        return response.data;
    }

    getOasByService = async (selectedService) => {
        const response = await client.get(`/oas/${selectedService}`);
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

    isInstallNeeded = async () => {
        const response = await client.get("/install/isNeeded");
        return response.data;
    }

    isAuthenticationNeeded = async () => {
        try {
            await client.get("/user/whoAmI");
            return false;
        } catch (e) {
            if (e.response.status === 401) {
                return true;
            }
            throw e;
        }
    }

    whoAmI = async () => {
        const response = await client.get("/user/whoAmI");
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

export function getToken(){
    return token
}
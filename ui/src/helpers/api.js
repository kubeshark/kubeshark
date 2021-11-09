import * as axios from "axios";

// When working locally cp `cp .env.example .env`
export const MizuWebsocketURL = process.env.REACT_APP_OVERRIDE_WS_URL ? process.env.REACT_APP_OVERRIDE_WS_URL : `ws://${window.location.host}/ws`;

export default class Api {

    constructor() {

        // When working locally cp `cp .env.example .env`
        const apiURL = process.env.REACT_APP_OVERRIDE_API_URL ? process.env.REACT_APP_OVERRIDE_API_URL : `${window.location.origin}/`;

        this.client = axios.create({
            baseURL: apiURL,
            timeout: 31000,
            headers: {
                Accept: "application/json",
            }
        });
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

    getRecentTLSLinks = async () => {
        const response = await this.client.get("/status/recentTLSLinks");
        return response.data;
    }

    getAuthStatus = async () => {
        const response = await this.client.get("/status/auth");
        return response.data;
    }

    validateQuery = async (query) => {
        const form = new FormData();
        form.append('query', query)
        const response = await this.client.post(`/query/validate`, form);
        return response.data;
    }
}

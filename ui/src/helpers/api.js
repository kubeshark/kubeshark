import * as axios from "axios";

const mizuAPIPathPrefix = "/mizu";

// When working locally (with npm run start) change to:
// export const MizuWebsocketURL = `ws://localhost:8899${mizuAPIPathPrefix}/ws`;
export const MizuWebsocketURL = `ws://${window.location.host}${mizuAPIPathPrefix}/ws`;

export default class Api {
    constructor() {
        this.client = null;
        // When working locally (with npm run start) change to:
        // this.api_url = `http://localhost:8899/${mizuAPIPathPrefix}/api/`;
        this.api_url = `${window.location.origin}${mizuAPIPathPrefix}/api/`;
    }

    init = () => {

        let headers = {
            Accept: "application/json",
        };

        this.client = axios.create({
            baseURL: this.api_url,
            timeout: 31000,
            headers: headers,
        });

        return this.client;
    };

    tapStatus = async () => {
        const response = await this.init().get("/tapStatus");
        return response.data;
    }

    analyzeStatus = async () => {
        const response = await this.init().get("/analyzeStatus");
        return response.data;
    }

    getEntry = async (entryId) => {
        const response = await this.init().get(`/entries/${entryId}`);
        return response.data;
    }

    fetchEntries = async (operator, timestamp) => {
        const response = await this.init().get(`/entries?limit=50&operator=${operator}&timestamp=${timestamp}`);
        return response.data;
    }
}
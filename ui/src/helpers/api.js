import * as axios from "axios";

const mizuAPIPathPrefix = "/mizu";

// When working locally (with npm run start) change to:
// export const MizuWebsocketURL = `ws://localhost:8899/ws`;
export const MizuWebsocketURL = `ws://${window.location.host}${mizuAPIPathPrefix}/ws`;

export default class Api {

    constructor() {

        // When working locally (with npm run start) change to:
        // const apiURL = `http://localhost:8899/api/`;
        const apiURL = `${window.location.origin}${mizuAPIPathPrefix}/api/`;

        this.client = axios.create({
            baseURL: apiURL,
            timeout: 31000,
            headers: {
                Accept: "application/json",
            }
        });
    }

    tapStatus = async () => {
        const response = await this.client.get("/tapStatus");
        return response.data;
    }

    analyzeStatus = async () => {
        const response = await this.client.get("/analyzeStatus");
        return response.data;
    }

    getEntry = async (entryId) => {
        const response = await this.client.get(`/entries/${entryId}`);
        return response.data;
    }

    fetchEntries = async (operator, timestamp) => {
        const response = await this.client.get(`/entries?limit=50&operator=${operator}&timestamp=${timestamp}`);
        return response.data;
    }

    getRecentTLSLinks = async () => {
        const response = await this.client.get("/recentTLSLinks");
        return response.data;
    }
}

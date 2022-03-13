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
        const response = await this.client.post("/config/tap", {tappedNamespaces: config});
        return response.data;
    }

    //#region User api

    getUsers = async(filter = "") =>{
        const response = await client.get(`/user/listUsers?usernameFilter=${filter}`);
        return response.data;
    }

    getUserDetails = async(userId) => {
        const response = await client.get(`/user/${userId}`);
        return response.data;
    }

    updateUser = async(user) => {
        const response = await client.put(`/user/${user.userId}`,user);
        return response.data;
    }

    deleteUser = async(userId) => {
        const response = await client.delete(`/user/${userId}`);
        return response.data;
    }

    recoverUser = async(data) => {
        const form = new FormData();
        form.append('password', data.password);
        form.append('inviteToken', data.inviteToken);

        try {
            const response = await client.post(`/user/recover`, form);
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

    inviteExistingUser = async(userId)  =>{
        const response = await client.post(`/user/${userId}/invite`);
        return response.data;
    }

     genareteInviteLink = async(userData)  =>{
        const response = await client.post(`/user/createUserAndInvite`,userData);
        return response.data;
    }

    //#endregion 

    //#region Workspace api
    getWorkspaces = async() =>{
        const response = await client.get(`/workspace/`);
        return response.data;
    }

    getSpecificWorkspace = async(workspaceId) =>{
        const response = await client.get(`/workspace/${workspaceId}`);
        return response.data;
    }

    createWorkspace = async(workspaceData,linkUser) =>{
        let path = `/workspace/`;
        if(linkUser){
            path = `/workspace/?linkUser=${linkUser}`;
        }
        const response = await client.post(path,workspaceData);
        return response.data;
    }    

    editWorkspace = async(workspaceId, workspaceData) =>{
        const response = await client.put(`/workspace/${workspaceId}`,workspaceData);
        return response.data;
    }   

    deleteWorkspace = async(workspaceId) => {
        const response = await client.delete(`/workspace/${workspaceId}`);
        return response.data;
    }

    getNamespaces = async() =>{
        const response = await client.get(`/kube/namespaces`);
        return response.data;
    }

    //#endregion

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

    getServerConfig = async () => {
        const response = await client.get("/config/");
        return response.data;
    }

    saveServerConfig = async (newConfig) => {
        const response = await client.post("/config/", newConfig);
        return response.data;
    }

    getDefaultServerConfig = async () => {
        const response = await client.get("/config/defaults");
        return response.data;
    }

    getServerMetadata = async () => {
        const response = await client.get("/metadata/");
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

    setupAdminUser = async (password) => {
        const form = new FormData();
        form.append('password', password);

        try {
            const response = await client.post(`/install/admin`, form);
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

        const response = await client.post(`/user/login`, form);
        if (response.status >= 200 && response.status < 300) {
            this.persistToken(response.data.token);
        }

        return response;
    }

    persistToken = (tk) => {
        token = tk;
        client = this.getAxiosClient();
        localStorage.setItem('token', token);
    }

    logout = async () => {
        await client.post(`/user/logout`);
        this.persistToken(null);
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

export function getWebsocketUrl(){
    let websocketUrl = MizuWebsocketURL;
    if (token) {
      websocketUrl += `/${token}`;
    }

    return websocketUrl;
}
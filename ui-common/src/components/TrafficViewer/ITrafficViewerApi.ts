type ITrafficViewerApi = {
    validateQuery : (query: any) => any
    tapStatus : () => any
    analyzeStatus : () => any
    fetchEntries : (leftOff: any, direction: number, query: any, limit: number, timeoutMs: number) => any
    getEntry : (entryId : any) => any
    getRecentTLSLinks : () => any
  }

  export default ITrafficViewerApi
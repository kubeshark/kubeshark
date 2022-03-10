type TrafficViewerApi = {
    validateQuery : (query: any) => any
    tapStatus : () => any
    analyzeStatus : () => any
    fetchEntries : (leftOff: any, direction: number, query: any, limit: number, timeoutMs: number) => any
    getEntry : (entryId : any, query:string) => any
    getRecentTLSLinks : () => any
  }

  export default TrafficViewerApi
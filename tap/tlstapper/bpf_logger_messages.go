package tlstapper

// Must be synced with logger_messages.h
//
var bpfLogMessages = []string {
	/*0000*/ "[%d] Unable to read bytes count from _ex methods [err: %d]",
	/*0001*/ "[%d] Unable to read ipv4 address [err: %d]",
	/*0002*/ "[%d] Unable to read ssl buffer [err: %d]",
	/*0003*/ "[%d] Buffer is too big [size: %d]",
	/*0004*/ "[%d] Unable to allocate chunk in bpf heap",
	/*0005*/ "[%d] Unable to read ssl context [err: %d] [origin: %d]",
	/*0006*/ "[%d] Unable to put ssl context [err: %d]",
	/*0007*/ "[%d] Unable to get ssl context",
	/*0008*/ "[%d] File descriptor is missing for tls chunk",
	/*0009*/ "[%d] Unable to put file descriptor [err: %d] [origin: %d]",
	/*0010*/ "[%d] Unable to put accept info [err: %d]",
	/*0011*/ "[%d] Unable to get accept info",
	/*0012*/ "[%d] Unable to read accept info [err: %d]",
	/*0013*/ "[%d] Unable to put file descriptor to address mapping [err: %d] [origin: %d]",
	/*0014*/ "[%d] Unable to put connect info [err: %d]",
	/*0015*/ "[%d] Unable to get connect info",
	/*0016*/ "[%d] Unable to read connect info [err: %d]",
	
}


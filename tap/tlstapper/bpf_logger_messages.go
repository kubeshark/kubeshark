package tlstapper

// Must be synced with logger_messages.h
//
var bpfLogMessages = []string{
	/*0000*/ "[%d] Unable to read bytes count from _ex methods [err: %d]",
	/*0001*/ "[%d] Unable to read ssl buffer [err: %d] [origin: %d]",
	/*0002*/ "[%d] Buffer is too big [size: %d]",
	/*0003*/ "[%d] Unable to allocate chunk in bpf heap",
	/*0004*/ "[%d] Unable to read ssl context [err: %d] [origin: %d]",
	/*0005*/ "[%d] Unable to put ssl context [err: %d]",
	/*0006*/ "[%d] Unable to get ssl context",
	/*0007*/ "[%d] File descriptor is missing for tls chunk",
	/*0008*/ "[%d] Unable to put file descriptor [err: %d] [origin: %d]",
	/*0009*/ "[%d] Unable to put accept info [err: %d]",
	/*0010*/ "[%d] Unable to get accept info",
	/*0011*/ "[%d] Unable to read accept info [err: %d]",
	/*0012*/ "[%d] Unable to put connection info to connection context [err: %d] [origin: %d]",
	/*0013*/ "[%d] Unable to put connect info [err: %d]",
	/*0014*/ "[%d] Unable to get connect info",
	/*0015*/ "[%d] Unable to read connect info [err: %d]",
	/*0016*/ "[%d] Unable to read socket family [err: %d]",
	/*0017*/ "[%d] Unable to read socket daddr [err: %d]",
	/*0018*/ "[%d] Unable to read socket saddr [err: %d]",
	/*0019*/ "[%d] Unable to read socket dport [err: %d]",
	/*0020*/ "[%d] Unable to read socket sport [err: %d]",
	/*0021*/ "[%d] Unable to put go user-kernel context [fd: %d] [err: %d]",
	/*0022*/ "[%d] Unable to get go user-kernel context [fd: %d]]",
}

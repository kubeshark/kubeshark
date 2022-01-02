package oas

import "sync"

// TODO: should it go somewhere like "providers" package?

var ServiceSpecs = &sync.Map{}

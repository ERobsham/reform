# Reform

Reform your unstructured system logs into [CLEF](https://clef-json.org/) structured logs.

Use config files to easily setup multiple input and output streams.  For example:

```
$ reform -config=config.json -seq=localhost:5341
```

(example config.json contents)
``` json
{
    "sources":{
        "10.0.0.30":{
            "cmd": "ssh",
            "args":["user@10.0.0.30", "tail -F /var/log/system.log"]
        },
        "10.0.0.40":{
            "cmd": "ssh",
            "args":["user@10.0.0.40", "tail -F /var/log/syslog"]
        },
    },
    "outputs": {
        "stdout":{
            "type": "stdout"
        },
        "file-log":{
            "type": "file",
            "config":{
                "path":"test.log"
            }
        }
    }
}
```


Assuming you have permissions, that allows you to stream system logs from both `10.0.0.30` and `10.0.0.40`, while outputting the parsed structured logs as a shortened summary to `stdout`, as [CLEF](https://clef-json.org/) structured logs to `test.log`, and finally, the `-seq=localhost:5341` also pushes logs to a local instance of [Seq](https://datalust.co/seq) for a awesome UI to view / search / filter the structured logs.


### Note:

This tool is just a toy project I made for myself to make slogging through unstructured logs more pleasant.  

It parses using heuristics to decide if log lines should be concatenated together (multi-line JSON-like logs), or if it should consume some prefix / suffix from a log line.

As such, its tailored to the style of logs I see, and it may not properly parse everything you're looking properly -- although its aim is to _not_ parse something if its not within its set of heuristics, so unparsed chunks just end up as the main 'message'.

I hope its helpful!  If you find a class of logs where the parsing is lacking, feel free to add some tests make a PR.

### gossh
parallel ssh command execution in go


##### usage
```
Usage:
cat iplist | ./gossh [-w|--workers] [-u|--username] [-i|--identity] 'cmd to run'
  --workers  -w -- Number of workers to spawn (default: 3)
  --username -u -- Username to use for ssh connections (default: alex)
  --identity -i -- ssh private key to use (default: /home/alex/.ssh/id_rsa)

iplist must be \n delimited list
```



##### example
```
$ cat iplist
10.10.105.101
10.10.105.102
10.10.105.103
$ cat iplist | gossh -u arwineap "date"
10.10.105.101: Sun Mar 15 20:16:55 PDT 2015
10.10.105.102: Sun Mar 15 20:16:55 PDT 2015
10.10.105.103: Sun Mar 15 20:16:55 PDT 2015
```


## Linux Command GPT (lcg)
Get Linux commands in natural language with the power of ChatGPT.

### Installation
Build from source
```bash
> git clone --depth 1 https://github.com/asrul10/linux-command-gpt.git ~/.linux-command-gpt
> cd ~/.linux-command-gpt
> go build -o lcg
# Add to your environment $PATH
> ln -s ~/.linux-command-gpt/lcg ~/.local/bin
```

Or you can [download lcg executable file](https://github.com/asrul10/linux-command-gpt/releases)

### Example Usage

```bash
> lcg I want to extract linux-command-gpt.tar.gz file
Completed in 0.92 seconds

tar -xvzf linux-command-gpt.tar.gz 

Do you want to (c)opy, (r)egenerate, or take (N)o action on the command? (c/r/N):
```

```bash
> LCG_PROMPT='Provide full response' LCG_MODEL=codellama:13b lcg 'i need bash script 
to execute some command by ssh on some array of hosts'
Completed in 181.16 seconds

Here is a sample Bash script that demonstrates how to execute commands over SSH on an array of hosts:
```bash
#!/bin/bash

hosts=(host1 host2 host3)

for host in "${hosts[@]}"; do
  ssh $host "echo 'Hello, world!' > /tmp/hello.txt"
done
```
This script defines an array `hosts` that contains the names of the hosts to connect to. The loop iterates over each element in the array and uses the `ssh` command to execute a simple command on the remote host. In this case, the command is `echo 'Hello, world!' > /tmp/hello.txt`, which writes the string "Hello, world!" to a file called `/tmp/hello.txt`.

You can modify the script to run any command you like by replacing the `echo` command with your desired command. For example, if you want to run a Python script on each host, you could use the following command:
```bash
ssh $host "python /path/to/script.py"
```
This will execute the Python script located at `/path/to/script.py` on the remote host.

You can also modify the script to run multiple commands in a single SSH session by using the `&&` operator to chain the commands together. For example:
```bash
ssh $host "echo 'Hello, world!' > /tmp/hello.txt && python /path/to/script.py"
```
This will execute both the `echo` command and the Python script in a single SSH session.

I hope this helps! Let me know if you have any questions or need further assistance.

Do you want to (c)opy, (r)egenerate, or take (N)o action on the command? (c/r/N):
```

To use the "copy to clipboard" feature, you need to install either the `xclip` or `xsel` package.

### Options
```bash
> lcg [options]

--help        -h  output usage information
--version     -v  output the version number
--file        -f  read command from file
--update-key  -u  update the API key
--delete-key  -d  delete the API key
```

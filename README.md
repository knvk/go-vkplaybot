# go-vkplaybot
VK Play Live chat bot in Go.

Most of the code is written just for myself and fun, and does not pretend to be "ideologically correct"
It's not ready yet and should be considered as pre-pre-alpha.

## Install and usage:

```
go build -o bot .
```
Then run with:
```
bot /path/to/config
```
Example config file can be found [here](https://github.com/knvk/go-vkplaybot/blob/main/config.cfg)

### Some important notes
- Message processing in separate coroutine
- TOML formatted config
- Gorillas websocket wrapper
- Ban phrases file should contain one word per line. 
- Modules order matter
- Can't use public timeout/delete/ban methods yet
- If auth-token (complex one with refresh and expire values) provided it will be used. You can get it via `bot.auth` func 
  **OR** just from yours web-browser `devtools->application->cookies->vkplay.live`
 
## Commands

Some several commands example:
- Dick'o meter
- Greeter (just reply to command)
- Joke from anekdot.ru public api
- Follow Age (it uses boosty one, dunno why devs did this LUL)
- Viewers list (with mods, banned, owner)

## TODO

- Implement message manipulation methods
- Vote Poll module
- Points (with sqlite backend?)

Any feedback and suggestions are welcome

## Remark
This project is distributed as is and is not associated with the VK company and the VKPlay service. The repository does not contain the source code of the VKontakte nor VKPlay applications. This software uses available public interfaces and APIs.

## Support

You can support further work at [boosty](https://boosty.to/divanic)

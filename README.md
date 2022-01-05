## Configure the dev environment for Windows 10

### Install compiler

#### Option 1: msys2

Install mingw-w64 then msys2.

#### Option 2: TDM-GCC

TDM-GCC supports Windows XP or higher.

### Install GoLang

You can specify workspace of GoLang to other else C drive.
For example, `F:\go-workspace`.

## Troubleshooting

### Prevent Windows Defender Firewall from appearing on every launch

On build, GoLang creates the executable as temporary file with appending uuid to filename.
On every launch, new file requests the 8080 port, so Windows Defender Firewall displays the dialog prompt that asks whether to allow private or public network.
This is annoying, because it appears on every launch.

Add `127.0.0.1` to the calling of `server.Router.Run`, in addition to port.

`server.Router.Run("127.0.0.1:" + os.Getenv("PORT"))`

### Prevent Avast from scanning this executable on every launch

On build, GoLang creates the executable as temporary file with appending uuid to filename of `__debug_bin`.
Avast scans this new file in Windows temp directory.
This is annoying, because it appears on every launch.

Add `C:\Users\xxx\AppData\Local\Temp\__debug_bin*.exe` to exception list of Avast.

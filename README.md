**1pw** is a quick terminal-based viewer for 1password 'opvault' data.

1pw doesn't have anything to do with the agilebits subscription/sync service. I don't know if it's possible to use. It does work with dropbox or manually synced vaults.

## Usage

You need [https://github.com/junegunn/fzf](fzf) and `xclip`. Use the usual `go build` / `go install` sort of process to build `1pw`.

Symlink your vault to `~/.1pw` or specify the path with `-vault path`.

After prompting for the vault password, you'll see a FZF search over all items. Type to fuzzy search and select an item.

All information for the item is printed, with passwords censored (unless `-reveal` is used). The first password in the output will be automatically copied to the clipboard, unless `-no-copy` is passed.

Contributions welcome!

